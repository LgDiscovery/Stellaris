import numpy as np
from flask import Flask, request, jsonify
from pyexpat import features
from sklearn.datasets import load_iris
from sklearn.linear_model import LogisticRegression

#初始化 Flask应用

app = Flask(__name__)

# 训练一个简单的鸢尾花分类模型（实际场景中可加载预训练模型）
def train_model():
    iris = load_iris()
    X, y = iris.data, iris.target
    model = LogisticRegression()
    model.fit(X,y)
    return model,iris.target_names

# 加载模型和类别名称
model, target_names = train_model()

# 3. 定义推理接口（POST 请求）
@app.route('/predict', methods=['POST'])
def predict():
    try:
         # 获取请求中的特征数据（JSON 格式）
         data = request.get_json()
         if not data or 'features' not in data:
             return jsonify({'error': '请提供特征数据，格式：{"features": [5.1, 3.5, 1.4, 0.2]}'}), 400

         # 转换为模型可接受的格式
         features = np.array(data['features']).reshape(1, -1)

         # 模型推理
         pred_idx = model.predict(features)[0]
         pred_name = target_names[pred_idx]

         # 返回结果
         return jsonify({
             'prediction_index': int(pred_idx),
             'prediction_name': pred_name,
             'features': data['features']
         })
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 4. 启动服务
if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=5000)