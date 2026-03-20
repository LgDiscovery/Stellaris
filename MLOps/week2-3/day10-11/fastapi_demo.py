import uvicorn
from fastapi import FastAPI
from pydantic import BaseModel,Field,validator
from typing import Optional

app = FastAPI(title='My FastAPI app',version='0.0.1')

# 定义请求参数
class NumberRequest(BaseModel):
    num1: float = Field(...,description="第一个数字(必填)",gt=-1000,it=1000)
    num2: float = Field(...,description="第二个数字(必填)",gt=-1000,it=1000)
    request_id: Optional[str] = Field(None,description="请求ID(可选)",min_length=1,max_length=100)

@validator('num1','num2')
def not_zero_sum(cls,v,values):
    """自定义校验：禁止num1和num2同时为0（示例）"""
    # values包含已校验的字段（注意字段顺序，num1先校验时values无num2）
    if 'num1' in values and 'num2' in values:
        if values['num1'] == 0 and values['num2'] == 0:
            raise ValueError("num1和num2不能同时为0")
    return v

# 定义响应参数
class SumResponse(BaseModel):
    code: int = 200
    message: str = "SUCCESS"
    data: dict

@app.get("/")
async def root():
    return {"message": "Hello World"}

@app.post("/calculate/sum",response_model=SumResponse,summary="带请求体校验的数字求和接口")
async def calculate_sum(request: NumberRequest):
    sum_result = request.num1 + request.num2
    return {
        "code": 200,
        "message": "求和成功",
        "data": {
            "num1": request.num1,
            "num2": request.num2,
            "sum": sum_result,
            "request_id": request.request_id
        }
    }

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)