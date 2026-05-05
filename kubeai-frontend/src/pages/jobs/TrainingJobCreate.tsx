import { useState } from 'react';
import { Card, Form, Input, Select, Button, Typography, message, InputNumber, Switch } from 'antd';
import { useNavigate } from 'react-router-dom';
import { ArrowLeftOutlined, SendOutlined } from '@ant-design/icons';
import { jobsApi, SubmitTrainingReq } from '../../api/jobs';

const { Title } = Typography;

export default function TrainingJobCreatePage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const generateRequestId = () => `train-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const data: SubmitTrainingReq = {
        request_id: generateRequestId(),
        name: values.name,
        model_name: values.model_name || undefined,
        framework: values.framework,
        image: values.image,
        command: values.command ? values.command.split(' ').filter(Boolean) : [],
        args: values.args ? values.args.split(' ').filter(Boolean) : [],
        resources: {
          cpu: values.cpu || undefined,
          memory: values.memory || undefined,
          gpu: values.gpu || undefined,
        },
        distributed: values.distributed || false,
        worker_num: values.worker_num || undefined,
        master_num: values.master_num || undefined,
        env: values.env || undefined,
        dataset_path: values.dataset_path || undefined,
        output_path: values.output_path || undefined,
        priority: values.priority || 5,
        max_retries: values.max_retries || 3,
        enable_monitor: values.enable_monitor ?? true,
        enable_logs: values.enable_logs ?? true,
      };
      await jobsApi.submitTraining(data);
      message.success('训练任务已提交');
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
      <Title level={4}>提交训练任务</Title>
      <Card style={{ maxWidth: 700 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}
          initialValues={{ priority: 5, max_retries: 3, enable_monitor: true, enable_logs: true }}>
          <Form.Item name="name" label="任务名称" rules={[{ required: true }]}>
            <Input placeholder="训练任务名称" />
          </Form.Item>
          <Form.Item name="model_name" label="关联模型（可选）">
            <Input placeholder="关联的模型名称" />
          </Form.Item>
          <Form.Item name="framework" label="训练框架" rules={[{ required: true }]}>
            <Select options={[
              { label: 'PyTorch', value: 'PyTorch' },
              { label: 'TensorFlow', value: 'TensorFlow' },
            ]} />
          </Form.Item>
          <Form.Item name="image" label="镜像" rules={[{ required: true }]}>
            <Input placeholder="例如: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime" />
          </Form.Item>
          <Form.Item name="command" label="命令">
            <Input placeholder="例如: python train.py（空格分隔）" />
          </Form.Item>
          <Form.Item name="args" label="参数">
            <Input placeholder="例如: --epochs 10 --batch-size 32（空格分隔）" />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16 }}>
            <Form.Item name="cpu" label="CPU">
              <Input placeholder="例如: 4" />
            </Form.Item>
            <Form.Item name="memory" label="内存">
              <Input placeholder="例如: 8Gi" />
            </Form.Item>
            <Form.Item name="gpu" label="GPU">
              <Input placeholder="例如: 1" />
            </Form.Item>
          </div>
          <Form.Item name="distributed" label="分布式训练" valuePropName="checked">
            <Switch />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Form.Item name="worker_num" label="Worker数量">
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="master_num" label="Master数量">
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <Form.Item name="dataset_path" label="数据集路径">
            <Input placeholder="MinIO或存储路径" />
          </Form.Item>
          <Form.Item name="output_path" label="输出路径">
            <Input placeholder="模型输出存储路径" />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16 }}>
            <Form.Item name="priority" label="优先级">
              <InputNumber min={1} max={10} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_retries" label="最大重试次数">
              <InputNumber min={0} max={10} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Form.Item name="enable_monitor" label="启用监控" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="enable_logs" label="启用日志" valuePropName="checked">
              <Switch />
            </Form.Item>
          </div>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} icon={<SendOutlined />} size="large">
              提交训练任务
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}