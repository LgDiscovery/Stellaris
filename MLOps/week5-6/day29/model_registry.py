import mlflow
import numpy as np

# ====================== 1. 注册模型 ======================
# 把上一天的 run id 填进来（在 UI 里复制）
run_id = "94dc97ce0f95432f94ebf85e5874175d"
model_uri = f"runs:/{run_id}/model"

#注册模型
mlflow.register_model(
    model_uri=model_uri,
    name="My_LR_Model"  # 模型名
)

# ====================== 2. 切换模型阶段 ======================
client = mlflow.tracking.MlflowClient()

# 把 V1 设为 Production
client.transition_model_version_stage(
    name="My_LR_Model",
    version=1,
    stage="Production",
)

# ====================== 3. 加载模型并推理 ======================
model = mlflow.pyfunc.load_model("models:/My_LR_Model/Production")
# 测试推理
X_test = np.array([[6], [7]])
pred = model.predict(X_test)
print("预测结果：", pred)