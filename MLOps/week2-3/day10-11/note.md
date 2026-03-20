### 代码关键部分解释

1. **请求体模型（NumberRequest）**：
   * 继承`pydantic.BaseModel`，用于校验请求体数据类型，确保传入的`num1`和`num2`是数字（float 支持整数和小数）。
   * `request_id`是可选字段（`Optional`），演示请求体的可扩展性，适配模型推理中可能的额外参数（如请求 ID、模型版本等）。
2. **响应体模型（SumResponse）**：
   * 通过`response_model=SumResponse`指定接口返回格式，强制返回`code`、`message`、`data`三个字段，规范接口输出（适配模型推理的统一响应格式）。
3. **接口实现（calculate\_sum）**：
   * 路径为`/calculate/sum`，请求方法为`POST`（请求体推荐用 POST 方法传递）。
   * 函数参数`request: NumberRequest`会自动解析请求体，并校验数据类型，不符合则返回 422 错误。
   * 核心逻辑是计算两个数字的和，然后按规范格式返回结果。

### 调试步骤

1. **安装依赖**：

   先安装 FastAPI 和运行服务的 uvicorn：
   bash

   运行

   ```
   pip install fastapi uvicorn
   ```
2. **运行代码**：

   将代码保存为`sum_api.py`，执行命令：
   bash

   运行

   ```
   python sum_api.py
   ```

   服务会启动在`http://127.0.0.1:8000`。
3. **调试接口**：

   * **方式 1：FastAPI 自动文档（推荐）**：
     访问`http://127.0.0.1:8000/docs`，找到`/calculate/sum`接口，点击「Try it out」，输入请求体（示例）：json

     ```
     {
       "num1": 10.5,
       "num2": 20.3,
       "request_id": "req_123456"
     }
     ```

     点击「Execute」，即可看到返回结果：json

     ```
     {
       "code": 200,
       "message": "求和成功",
       "data": {
         "num1": 10.5,
         "num2": 20.3,
         "sum": 30.8,
         "request_id": "req_123456"
       }
     }
     ```
   * **方式 2：Postman/Curl**：
     发送 POST 请求到`http://127.0.0.1:8000/calculate/sum`，请求体为 JSON 格式，示例：bash

     运行

     ```
     curl -X POST "http://127.0.0.1:8000/calculate/sum" -H "Content-Type: application/json" -d '{"num1": 5, "num2": 3}'
     ```

     返回结果：json

     ```
     {"code":200,"message":"求和成功","data":{"num1":5.0,"num2":3.0,"sum":8.0,"request_id":null}}
     ```
