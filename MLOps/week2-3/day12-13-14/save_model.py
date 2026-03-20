import joblib
import numpy as np
import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.datasets import load_diabetes
from sklearn.model_selection import train_test_split

# ===================== 第一步：训练并保存预训练模型（模拟下载） =====================
# 加载经典的糖尿病数据集（用于训练线性回归模型）
data = load_diabetes()
X,y = data.data, data.target

# 划分训练集（无需测试集，仅为快速训练模型）
X_train, _, y_train, _ = train_test_split(X, y, test_size=0.2, random_state=42)

# 训练线性回归模型（模拟“预训练模型”）
model = LogisticRegression()
model.fit(X_train, y_train)

# 保存模型到本地文件（模拟“下载预训练模型”到本地）
model_path = "pretrained_linear_regression_model.joblib"
joblib.dump(model, model_path)
print(f"预训练模型已保存到：{model_path}")

# ===================== 第二步：加载预训练模型并推理 =====================
# 加载本地的预训练模型文件
loaded_model = joblib.load(model_path)
print("预训练模型加载完成！")
# 构造推理数据（需和训练时的特征维度一致，糖尿病数据集是10维特征）
# 这里随机构造1条测试数据，也可以用真实业务数据
inference_data = np.array([[0.03807591, 0.05068012, 0.06169621, 0.02187235, -0.0442235,
                            -0.03482076, -0.04340085, -0.00259226, 0.01990749, -0.01764613]])

# 执行推理（预测）
prediction = loaded_model.predict(inference_data)

# 输出结果
print(f"推理输入特征：{inference_data}")
print(f"模型预测结果（糖尿病病情指标）：{prediction[0]:.2f}")