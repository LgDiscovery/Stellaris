import { useState } from 'react';
import { Card, Form, Input, Select, Button, Typography, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { ArrowLeftOutlined, SendOutlined } from '@ant-design/icons';
import { jobsApi, SubmitInferenceReq } from '../../api/jobs';

const { Title } = Typography;
const { TextArea } = Input;

export default function InferenceJobCreatePage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const generateRequestId = () => `inf-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      let inputData: Record<string, unknown>;
      try {
        inputData = JSON.parse(values.input_data);
      } catch {
        message.error('输入数据格式错误，请输入有效的JSON');
        setLoading(false);
        return;
      }

      const data: SubmitInferenceReq = {
        request_id: generateRequestId(),
        model_name: values.model_name,
        model_version: values.model_version,
        resources: {
          cpu: values.cpu || undefined,
          memory: values.memory || undefined,
          gpu: values.gpu || undefined,
        },
        input_data: inputData,
        output_topic: values.output_topic || undefined,
        priority: values.priority || 5,
        max_retries: values.max_retries || 3,
      };
      await jobsApi.submitInference(data);
      message.success('推理任务已提交');
      navigate('/jobs');
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/jobs')} style={{ marginBottom: 16 }}>
        返回列表
      </Button>
      <Title level={4}>提交推理任务</Title>
      <Card style={{ maxWidth: 700 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}
          initialValues={{ priority: 5, max_retries: 3 }}>
          <Form.Item name="model_name" label="模型名称" rules={[{ required: true }]}>
            <Input placeholder="例如: resnet50" />
          </Form.Item>
          <Form.Item name="model_version" label="模型版本" rules={[{ required: true }]}>
            <Input placeholder="例如: v1.0.0" />
          </Form.Item>
          <Form.Item name="input_data" label="输入数据 (JSON)" rules={[{ required: true }]}>
            <TextArea rows={5} placeholder='{"image_url": "...", "params": {}}' />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16 }}>
            <Form.Item name="cpu" label="CPU资源">
              <Input placeholder="例如: 2" />
            </Form.Item>
            <Form.Item name="memory" label="内存资源">
              <Input placeholder="例如: 4Gi" />
            </Form.Item>
            <Form.Item name="gpu" label="GPU资源">
              <Input placeholder="例如: 1" />
            </Form.Item>
          </div>
          <Form.Item name="output_topic" label="输出Topic（可选）">
            <Input placeholder="结果回调topic" />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Form.Item name="priority" label="优先级">
              <Select options={[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map(n => ({ label: String(n), value: n }))} />
            </Form.Item>
            <Form.Item name="max_retries" label="最大重试次数">
              <Select options={[0, 1, 2, 3, 5, 10].map(n => ({ label: String(n), value: n }))} />
            </Form.Item>
          </div>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} icon={<SendOutlined />} size="large">
              提交推理任务
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}