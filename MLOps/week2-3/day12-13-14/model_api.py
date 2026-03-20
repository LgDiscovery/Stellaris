import joblib
import numpy as np
from fastapi import FastAPI
from pydantic import BaseModel

# ===================== 1. 初始化FastAPI应用 =====================
app = FastAPI(
    title="糖尿病预测API",
    description="基于scikit-learn线性回归模型的糖尿病病情预测接口",
    version = "1.0.0"
)


# ===================== 2. 加载预训练模型 =====================
# 替换为你之前保存的预训练模型文件路径
MODEL_PATH = "pretrained_linear_regression_model.joblib"
try:
    # 加载模型（启动服务时一次性加载，避免每次请求都加载）
    model = joblib.load(MODEL_PATH)
    print(f"✅ 预训练模型加载成功：{MODEL_PATH}")
except Exception as e:
    print(f"❌ 模型加载失败：{e}")
    raise e

# ===================== 3. 定义请求数据格式（Pydantic校验） =====================
# 糖尿病数据集有10个特征，这里定义每个特征的字段（名称仅为示例，可根据实际业务调整）
class DiabetesFeatures(BaseModel):
    feature1: float  # 对应数据集第1个特征
    feature2: float  # 对应数据集第2个特征
    feature3: float  # 对应数据集第3个特征
    feature4: float  # 对应数据集第4个特征
    feature5: float  # 对应数据集第5个特征
    feature6: float  # 对应数据集第6个特征
    feature7: float  # 对应数据集第7个特征
    feature8: float  # 对应数据集第8个特征
    feature9: float  # 对应数据集第9个特征
    feature10: float # 对应数据集第10个特征


# ===================== 4. 定义API接口 =====================
@app.post("/predict",summary="糖尿病病情预测")
async def predict(diabetes_features: DiabetesFeatures):
    """
    接收特征数据，返回模型推理结果
    - 请求体：10个浮点型特征
    - 返回：预测的糖尿病病情指标值
    """
    try:
        # 1. 把请求数据转换为模型需要的格式（二维数组）
        input_data = np.array([[
            diabetes_features.feature1,
            diabetes_features.feature2,
            diabetes_features.feature3,
            diabetes_features.feature4,
            diabetes_features.feature5,
            diabetes_features.feature6,
            diabetes_features.feature7,
            diabetes_features.feature8,
            diabetes_features.feature9,
            diabetes_features.feature10
        ]])

        # 2.模型推理
        prediction = model.predict(input_data)

        # 3.构造返回结果
        return {
            "code": 200,
            "message": "success",
            "data":{
                "input_features": input_data.tolist(),  # 转换为列表方便JSON序列化
                "prediction_result": round(float(prediction[0]), 2)  # 保留2位小数
            }
        }
    except Exception as e:
        # 异常处理
       return {
           "code": 500,
           "message": f"推理失败:{str(e)}",
           "data": None
       }

# ===================== 5. 启动服务（本地测试用） =====================
if __name__ == "__main__":
    import uvicorn
    # 启动服务：host=0.0.0.0 允许外部访问，port=8000 端口号
    uvicorn.run(
        app = "model_api:app",
        host = "0.0.0.0",
        port = 8000,
        reload = True # 开发模式：代码修改后自动重启服务
    )