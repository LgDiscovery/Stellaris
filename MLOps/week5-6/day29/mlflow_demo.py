import mlflow.sklearn
from sklearn.linear_model import LinearRegression
import numpy as np

# ====================== 1. 设置实验 ======================
mlflow.set_experiment("my_first_mlops_exp")

# ====================== 2. 开始一次 Run ======================
with mlflow.start_run(run_name="lr_model_v1"):
    # -------------- 3. 准备模拟数据 --------------
    X = np.array([[1], [2], [3], [4], [5]])
    y = np.array([2, 4, 6, 8, 10])

    # -------------- 4. 定义模型参数 --------------
    fit_intercept = True
    model = LinearRegression(fit_intercept=fit_intercept)

    # -------------- 5. 训练 --------------
    model.fit(X, y)
    score = model.score(X, y)

    # -------------- 6. 记录到 MLflow --------------
    # 记录参数
    mlflow.log_param("fit_intercept", fit_intercept)
    mlflow.log_param("model_type", "LinearRegression")

    # 记录指标
    mlflow.log_metric("R2_score", score)

    # 记录模型
    mlflow.sklearn.log_model(model, "model")

    print("训练完成！打开 http://localhost:5000 查看")
    print(f"模型R2分数：{score}")