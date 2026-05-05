import { useState } from 'react';
import { Card, Form, Input, Select, Button, Typography, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { modelsApi, CreateModelReq } from '../../api/models';

const { Title } = Typography;

export default function ModelCreatePage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const onFinish = async (values: CreateModelReq) => {
    setLoading(true);
    try {
      await modelsApi.create(values);
      message.success('模型创建成功');
      navigate('/models');
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/models')} style={{ marginBottom: 16 }}>
        返回列表
      </Button>
      <Title level={4}>创建模型</Title>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="name" label="模型名称" rules={[{ required: true, message: '请输入模型名称' }]}>
            <Input placeholder="例如: resnet50, bert-base" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="模型描述信息" />
          </Form.Item>
          <Form.Item name="framework" label="框架">
            <Select
              placeholder="选择框架"
              options={[
                { label: 'PyTorch', value: 'PyTorch' },
                { label: 'TensorFlow', value: 'TensorFlow' },
                { label: 'ONNX', value: 'ONNX' },
                { label: '其他', value: 'Other' },
              ]}
            />
          </Form.Item>
          <Form.Item name="task_type" label="任务类型">
            <Select
              placeholder="选择任务类型"
              options={[
                { label: '图像分类', value: 'image_classification' },
                { label: '目标检测', value: 'object_detection' },
                { label: '文本生成', value: 'text_generation' },
                { label: '语音识别', value: 'speech_recognition' },
                { label: '其他', value: 'other' },
              ]}
            />
          </Form.Item>
          <Form.Item name="owner" label="负责人">
            <Input placeholder="负责人用户名" />
          </Form.Item>
          <Form.Item name="labels" label="标签">
            <Input placeholder="逗号分隔的标签，如: production,stable" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              创建模型
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}