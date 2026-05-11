import logging
import numpy as np
import mlflow.pyfunc
from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel, Field, field_validator, ValidationInfo

# ===================== 1. 配置日志（方便调试） =====================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
)

logger = logging.getLogger("diabetes_api")

# ===================== 2. MLflow 模型配置（核心修改） =====================
# 你可以根据自己的 MLflow 环境修改这 3 个配置
MLFLOW_TRACKING_URI = "http://localhost:5000"  # MLflow 服务地址
MLFLOW_MODEL_NAME = "diabetes_linear_regression"  # 模型注册名称
MODEL_VERSION = "latest"  # 使用 latest / 1 / 2 / Production 等版本/阶段

# 连接 MLflow
mlflow.set_tracking_uri(MLFLOW_TRACKING_URI)
mlflow.set_registry_uri(MLFLOW_TRACKING_URI)

# ===================== 3. 初始化FastAPI应用 =====================
app = FastAPI(
    title="糖尿病预测API（MLflow版）",
    description="基于 MLflow 模型仓库加载线性回归模型，提供糖尿病病情预测服务",
    version="3.0.0"
)

# ===================== 4. 从 MLflow 加载模型（核心替换） =====================
try:
    logger.info(f"正在从 MLflow 加载模型：{MLFLOW_MODEL_NAME} 版本：{MODEL_VERSION}")

    # 方法1：加载 注册模型（推荐，支持版本/阶段）
    model = mlflow.pyfunc.load_model(
        model_uri=f"models:/{MLFLOW_MODEL_NAME}/{MODEL_VERSION}"
    )

    # 方法2：如果用 run id 加载（不推荐固定写死，仅备用）
    # model = mlflow.pyfunc.load_model("runs:/<run-id>/model")

    logger.info(f"✅ MLflow 模型加载成功：{MLFLOW_MODEL_NAME}@{MODEL_VERSION}")

except Exception as e:
    logger.error(f"❌ MLflow 模型加载失败：{str(e)}")
    raise RuntimeError(f"模型加载失败，请检查 MLflow 服务和模型名称")

# ===================== 5. 定义请求数据模型（无修改） =====================
class DiabetesFeatures(BaseModel):
    """
    糖尿病预测特征模型，包含严格的参数校验
    特征范围基于scikit-learn diabetes数据集的实际分布（-0.1 ~ 0.1）
    """
    feature1: float = Field(..., description="糖尿病数据集特征1", example=0.03807591)
    feature2: float = Field(..., description="特征2", example=0.05068012)
    feature3: float = Field(..., description="特征3", example=0.06169621)
    feature4: float = Field(..., description="特征4", example=0.02187235)
    feature5: float = Field(..., description="特征5", example=-0.0442235)
    feature6: float = Field(..., description="特征6", example=-0.03482076)
    feature7: float = Field(..., description="特征7", example=-0.04340085)
    feature8: float = Field(..., description="特征8", example=-0.00259226)
    feature9: float = Field(..., description="特征9", example=0.01990749)
    feature10: float = Field(..., description="特征10", example=-0.01764613)

    # 自定义校验器：限制所有特征值范围
    @field_validator(
        "feature1", "feature2", "feature3", "feature4", "feature5",
        "feature6", "feature7", "feature8", "feature9", "feature10"
    )
    def check_feature_range(cls, v: float, info: ValidationInfo) -> float:
        if not (-0.1 <= v <= 0.1):
            raise ValueError(
                f"特征 {info.field_name} 值 {v} 超出范围！请输入 -0.1 ~ 0.1"
            )
        return v

# ===================== 6. 统一异常处理（无修改） =====================
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request, exc):
    logger.error(f"参数校验失败：{exc.errors()}")
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "code": 422,
            "message": "请求参数错误",
            "details": exc.errors(),
            "data": None
        }
    )

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

# ===================== 7. 核心推理接口（MLflow 兼容版） =====================
@app.post("/predict", summary="糖尿病病情预测", response_description="预测结果")
async def predict_diabetes(features: DiabetesFeatures):
    try:
        # 构造输入（MLflow pyfunc 支持 list/np.array）
        input_data = [
            [
                features.feature1, features.feature2, features.feature3,
                features.feature4, features.feature5, features.feature6,
                features.feature7, features.feature8, features.feature9,
                features.feature10
            ]
        ]
        logger.info(f"推理请求特征：{input_data}")

        # MLflow 模型预测（通用接口，无需关心模型框架）
        prediction = model.predict(input_data)
        result = round(float(prediction[0]), 2)

        return {
            "code": 200,
            "message": "推理成功",
            "data": {
                "input_features": input_data,
                "prediction_result": result,
                "model_version": MODEL_VERSION,
                "mlflow_model": MLFLOW_MODEL_NAME
            }
        }

    except Exception as e:
        logger.error(f"推理失败：{str(e)}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"模型推理失败：{str(e)}"
        )

# ===================== 8. 健康检查（增加 MLflow 信息） =====================
@app.get("/health", summary="服务健康检查", tags=["监控"])
async def health_check():
    try:
        # 测试推理
        test_data = np.zeros((1, 10)).tolist()
        model.predict(test_data)

        return {
            "code": 200,
            "message": "服务正常",
            "data": {
                "mlflow_tracking_uri": MLFLOW_TRACKING_URI,
                "model_name": MLFLOW_MODEL_NAME,
                "model_version": MODEL_VERSION,
                "api_version": "3.0.0"
            }
        }
    except Exception as e:
        return {
            "code": 500,
            "message": "服务异常",
            "data": {"error": str(e)}
        }

# ===================== 9. 启动服务 =====================
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        app="model_api_mlflow:app",  # 注意：文件名要对应！
        host="0.0.0.0",
        port=9000,
        reload=True
    )