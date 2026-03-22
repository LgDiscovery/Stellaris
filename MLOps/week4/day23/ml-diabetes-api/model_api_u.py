import logging
import joblib
import numpy as np
from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel, Field, field_validator,ValidationInfo

# ===================== 1. 配置日志（方便调试） =====================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
)

logger = logging.getLogger("diabetes_api")

# ===================== 2. 初始化FastAPI应用 =====================
app = FastAPI(
    title="糖尿病预测API（优化版）",
    description="基于scikit-learn线性回归模型的糖尿病病情预测接口，增加参数校验和异常处理",
    version="2.0.0"
)

# ===================== 3. 加载预训练模型（启动时加载，提升性能） =====================
MODEL_PATH = "pretrained_linear_regression_model.joblib"
try:
    model = joblib.load(MODEL_PATH)
    logger.info(f"✅ 预训练模型加载成功：{MODEL_PATH}")
except Exception as e:
    logger.error(f"❌ 模型加载失败：{e}")
    raise RuntimeError(f"模型加载失败，请检查文件路径：{MODEL_PATH}")

# ===================== 4. 定义请求数据模型（增强校验） =====================
class DiabetesFeatures(BaseModel):
    """
    糖尿病预测特征模型，包含严格的参数校验
    特征范围基于scikit-learn diabetes数据集的实际分布（-0.1 ~ 0.1）
    """
    feature1: float = Field(
        ...,  # 必填项
        description="糖尿病数据集特征1",
        example=0.03807591
    )
    feature2: float = Field(..., description="特征2", example=0.05068012)
    feature3: float = Field(..., description="特征3", example=0.06169621)
    feature4: float = Field(..., description="特征4", example=0.02187235)
    feature5: float = Field(..., description="特征5", example=-0.0442235)
    feature6: float = Field(..., description="特征6", example=-0.03482076)
    feature7: float = Field(..., description="特征7", example=-0.04340085)
    feature8: float = Field(..., description="特征8", example=-0.00259226)
    feature9: float = Field(..., description="特征9", example=0.01990749)
    feature10: float = Field(..., description="特征10", example=-0.01764613)

    # 自定义校验器：限制所有特征值的范围（-0.1 ~ 0.1）
    @field_validator("feature1", "feature2", "feature3", "feature4", "feature5",
                     "feature6", "feature7", "feature8", "feature9", "feature10")
    def check_feature_range(cls, v: float, info: ValidationInfo) -> float:
        """校验单个特征值是否在合法范围"""
        if not (-0.1 <= v <= 0.1):
            raise ValueError(
                f"特征 {info.field_name} 的值 {v} 超出合法范围！"
                f"请输入 -0.1 到 0.1 之间的数值"
            )
        return v

    # ===================== 5. 统一异常处理 =====================
# 处理请求参数校验失败的异常
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request, exc):
    logger.error(f"请求参数校验失败：{exc.errors()}")
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "code": 422,
            "message": "请求参数错误",
            "details": exc.errors(),  # 详细的错误信息
            "data": None
        }
    )

# 处理其他HTTP异常
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc):
    logger.error(f"HTTP异常：{exc.status_code} - {exc.detail}")
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "code": exc.status_code,
            "message": exc.detail,
            "data": None
        }
    )

# ===================== 6. 核心推理接口（优化版） =====================
@app.post("/predict", summary="糖尿病病情预测", response_description="预测结果")
async def predict_diabetes(features: DiabetesFeatures):
    """
    接收特征数据，返回模型推理结果
    - **输入**：10个浮点型特征（范围：-0.1 ~ 0.1）
    - **输出**：预测的糖尿病病情指标值
    """
    try:
        # 1. 转换请求数据为模型输入格式（二维数组）
        input_data = np.array([[
            features.feature1, features.feature2, features.feature3,
            features.feature4, features.feature5, features.feature6,
            features.feature7, features.feature8, features.feature9,
            features.feature10
        ]])
        logger.info(f"接收推理请求，输入特征：{input_data.tolist()}")

        # 2. 模型推理（异步函数中执行同步推理，不影响性能）
        prediction = model.predict(input_data)
        result = round(float(prediction[0]), 2)

        # 3. 返回结构化结果
        return {
            "code": 200,
            "message": "推理成功",
            "data": {
                "input_features": input_data.tolist(),
                "prediction_result": result
            }
        }

    except Exception as e:
        logger.error(f"推理过程异常：{str(e)}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"推理失败：{str(e)}"
        )

# ===================== 7. 健康检查接口（新增，方便监控） =====================
@app.get("/health", summary="服务健康检查", tags=["监控"])
async def health_check():
    """检查API服务和模型是否正常"""
    try:
        # 测试模型是否能正常推理（用空数据占位）
        test_data = np.zeros((1, 10))
        model.predict(test_data)
        return {
            "code": 200,
            "message": "服务正常",
            "data": {
                "model_loaded": True,
                "api_version": "2.0.0"
            }
        }
    except Exception as e:
        return {
            "code": 500,
            "message": "服务异常",
            "data": {
                "model_loaded": False,
                "error": str(e)
            }
        }

# ===================== 8. 启动服务 =====================
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        app="model_api_u:app",
        host="0.0.0.0",
        port=9000,
        reload=True,  # 开发模式，生产环境关闭
        log_level="info"
    )