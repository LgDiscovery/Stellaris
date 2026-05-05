import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Form, Input, Select, Button, Typography, message } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { modelsApi } from '../../api/models';

const { Title } = Typography;

export default function VersionCreatePage() {
  const { name } = useParams<{ name: string }>();
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const onFinish = async (values: {
    version: string;
    description?: string;
    framework?: string;
    framework_version?: string;
    metrics?: string;
    parameters?: string;
  }) => {
    if (!name) return;
    setLoading(true);
    try {
      await modelsApi.createVersion(name, { ...values, name });
      message.success('版本创建成功');
      navigate(`/models/${name}`);
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/models/${name}`)} style={{ marginBottom: 16 }}>
        返回详情
      </Button>
      <Title level={4}>为模型 "{name}" 添加版本</Title>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="version" label="版本号" rules={[{ required: true, message: '请输入版本号' }]}>
            <Input placeholder="例如: v1.0.0, 2.0.1" />
          </Form.Item>
          <Form.Item name="description" label="版本描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="framework" label="框架">
            <Select
              placeholder="选择框架"
              options={[
                { label: 'PyTorch', value: 'PyTorch' },
                { label: 'TensorFlow', value: 'TensorFlow' },
                { label: 'ONNX', value: 'ONNX' },
              ]}
            />
          </Form.Item>
          <Form.Item name="framework_version" label="框架版本">
            <Input placeholder="例如: 2.1.0" />
          </Form.Item>
          <Form.Item name="metrics" label="评估指标(JSON)">
            <Input.TextArea rows={2} placeholder='{"accuracy": 0.95, "loss": 0.12}' />
          </Form.Item>
          <Form.Item name="parameters" label="参数信息(JSON)">
            <Input.TextArea rows={2} placeholder='{"learning_rate": 0.001}' />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              创建版本
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}