import { useState } from 'react';
import { Card, Form, Input, Select, Button, Typography, message, Descriptions, Tag, Divider } from 'antd';
import { ThunderboltOutlined, SendOutlined } from '@ant-design/icons';
import { inferenceApi, InferenceRequest, InferenceResponse } from '../../api/inference';

const { Title, Text } = Typography;

export default function InferenceExecutePage() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<InferenceResponse | null>(null);
  const [form] = Form.useForm();

  const onFinish = async (values: InferenceRequest) => {
    setLoading(true);
    setResult(null);
    try {
      let input: Record<string, unknown>;
      try {
        input = typeof values.input === 'string' ? JSON.parse(values.input as unknown as string) : values.input;
      } catch {
        message.error('输入数据格式错误，请输入有效的JSON');
        setLoading(false);
        return;
      }
      const res = await inferenceApi.execute({ ...values, input });
      setResult(res.data);
      message.success('推理执行成功');
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={4}>
        <ThunderboltOutlined style={{ marginRight: 8 }} />
        执行推理
      </Title>
      <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
        提交推理请求到推理网关，获取模型推理结果
      </Text>
      <Card style={{ maxWidth: 700 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true, message: '请输入模型名称' }]}>
            <Input placeholder="例如: resnet50, bert-base" />
          </Form.Item>
          <Form.Item name="model_version" label="模型版本">
            <Input placeholder="例如: v1.0.0（可选）" />
          </Form.Item>
          <Form.Item name="framework" label="推理框架">
            <Select
              placeholder="选择框架（可选）"
              allowClear
              options={[
                { label: 'PyTorch', value: 'PyTorch' },
                { label: 'TensorFlow', value: 'TensorFlow' },
                { label: 'ONNX', value: 'ONNX' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="input"
            label="推理输入数据 (JSON)"
            rules={[{ required: true, message: '请输入推理数据' }]}
          >
            <Input.TextArea
              rows={6}
              placeholder='{"image_url": "https://example.com/image.jpg", "top_k": 5}'
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} icon={<SendOutlined />} size="large">
              执行推理
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {result && (
        <>
          <Divider />
          <Card title="推理结果" style={{ maxWidth: 700 }}>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="任务ID">
                <Tag color="blue">{result.task_id}</Tag>
              </Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16 }}>
              <Text strong>推理输出：</Text>
              <pre
                style={{
                  background: '#f5f5f5',
                  padding: 16,
                  borderRadius: 6,
                  marginTop: 8,
                  overflow: 'auto',
                  maxHeight: 400,
                  fontSize: 13,
                }}
              >
                {JSON.stringify(result.result, null, 2)}
              </pre>
            </div>
          </Card>
        </>
      )}
    </div>
  );
}