import os
import mlflow
import mlflow.sklearn
import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.datasets import load_diabetes
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error
from dotenv import load_dotenv

# 加载环境变量
load_dotenv(dotenv_path="../config/.env")

# ===================== 配置 MLflow 对接 MinIO =====================
# 给程序设置系统变量 MLflow 必须从环境变量读取 MinIO 地址，所以必须这么写
os.environ["MLFLOW_S3_ENDPOINT_URL"] = os.getenv("MINIO_ENDPOINT")
os.environ["AWS_ACCESS_KEY_ID"] = os.getenv("MINIO_ACCESS_KEY")
os.environ["AWS_SECRET_ACCESS_KEY"] = os.getenv("MINIO_SECRET_KEY")
mlflow.set_tracking_uri(os.getenv("MLFLOW_TRACKING_URI"))
mlflow.set_experiment("diabetes_prediction_lg_demo")

# ===================== 训练模型 =====================
if __name__ == "__main__":
    # 加载数据
    X, y = load_diabetes(return_X_y=True)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    with mlflow.start_run(run_name="linear_regression_baseline"):
        # 训练模型
        model = LinearRegression()
        model.fit(X_train, y_train)

        # 评估
        y_pred = model.predict(X_test)
        mse = mean_squared_error(y_test, y_pred)
        rmse = np.sqrt(mse)

        # MLflow 自动记录参数/指标/模型
        mlflow.log_param("model_type", "LinearRegression")
        mlflow.log_metric("mse", mse)
        mlflow.log_metric("rmse", rmse)

        # 保存模型到 MinIO（自动上传）
        mlflow.sklearn.log_model(
            sk_model=model,
            artifact_path="model",
            registered_model_name=os.getenv("MLFLOW_MODEL_NAME")  # 注册模型
        )

        print(f"✅ 模型训练完成，RMSE: {rmse:.2f}")
        print(f"✅ 模型已保存到 MinIO 并注册到 MLflow")