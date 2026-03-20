import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, classification_report
from sklearn.model_selection import train_test_split

# “数据处理→特征工程→模型训练→评估” 流程
# ====================== 1. Pandas：结构化数据处理（MLOps数据接入核心） ======================
# 模拟MLOps场景：读取业务数据（CSV/数据库是MLOps常见数据源）
data = pd.DataFrame({
    "age": [25, 30, 35, 40, 45, 50, 55, 60],
    "income": [5000, 8000, 10000, 15000, 12000, 20000, 18000, 25000],
    "credit_score": [650, 700, 680, 750, 720, 800, 780, 850],
    "default": [0, 0, 1, 0, 1, 0, 0, 1]  # 目标变量：是否违约（分类任务）
})

#MLOps必备：数据清洗 (处理缺失值，异常值)
data = data.dropna() # 删除缺失值（MLOps中需记录清洗规则）
data = data[data["income"] > 0]  # 过滤异常值

# 分离特征和目标变量
X = data[["age","income","credit_score"]]
y = data["default"]

# ====================== 2. NumPy：数值计算（MLOps特征矩阵处理） ======================
# 转换为NumPy数组（sklearn模型仅接受数组输入）
X_np = X.values  # 特征矩阵：shape (8, 3)
y_np = y.values  # 目标数组：shape (8,)

# MLOps常用：特征归一化（NumPy实现，也可通过sklearn简化）
X_np_scaled = (X_np - X_np.mean(axis=0)) / X_np.std(axis=0)

# ====================== 3. Scikit-learn：模型工程（MLOps核心环节） ======================
# 步骤1：划分训练集/测试集（MLOps中需固定随机种子，保证可复现）
X_train, X_test, y_train, y_test = train_test_split(
    X_np_scaled, y_np, test_size=0.2, random_state=42  # random_state保证结果可复现
)

# 步骤2：模型训练（选择逻辑回归，MLOps中常用基线模型）
model = LogisticRegression()
model.fit(X_train, y_train)

# 步骤3：模型评估（MLOps中需输出评估指标，用于监控）
y_pred = model.predict(X_test)
accuracy = accuracy_score(y_test, y_pred)
report = classification_report(y_test, y_pred)

print(f"模型准确率：{accuracy:.2f}")
print("分类报告：\n", report)

# MLOps延伸：保存模型（后续部署用）
import joblib
joblib.dump(model, "credit_default_model.pkl")  # 保存模型文件