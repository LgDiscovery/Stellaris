import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card, Descriptions, Table, Tag, Button, Space, Typography, Spin, message,
  Popconfirm, Modal, Form, Input, Select,
} from 'antd';
import {
  ArrowLeftOutlined, PlusOutlined, ReloadOutlined, DeleteOutlined, EditOutlined,
  DownloadOutlined, EyeOutlined,
} from '@ant-design/icons';
import { modelsApi, Model, ModelVersion } from '../../api/models';

const { Title } = Typography;

export default function ModelDetailPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const [model, setModel] = useState<Model | null>(null);
  const [versions, setVersions] = useState<ModelVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [editVisible, setEditVisible] = useState(false);
  const [editForm] = Form.useForm();

  const fetchData = async () => {
    if (!name) return;
    setLoading(true);
    try {
      const [modelRes, versionsRes] = await Promise.all([
        modelsApi.getByName(name),
        modelsApi.listVersions(name),
      ]);
      setModel(modelRes.data);
      setVersions(versionsRes.data.items || []);
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [name]);

  const handleEdit = async (values: { name: string; description?: string; labels?: string }) => {
    if (!name) return;
    try {
      await modelsApi.update(name, values);
      message.success('模型已更新');
      setEditVisible(false);
      fetchData();
    } catch {
      // 错误已处理
    }
  };

  const handleDeleteVersion = async (version: string) => {
    if (!name) return;
    try {
      await modelsApi.deleteVersion(name, version);
      message.success('版本已删除');
      fetchData();
    } catch {
      // 错误已处理
    }
  };

  const handleDownload = async (version: string) => {
    if (!name) return;
    try {
      const res = await modelsApi.downloadVersion(name, version);
      message.success(`下载路径: ${res.data.storage_path}`);
    } catch {
      // 错误已处理
    }
  };

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 60 }}><Spin size="large" /></div>;
  }

  if (!model) {
    return <div>模型不存在</div>;
  }

  const versionColumns = [
    { title: '版本号', dataIndex: 'version', key: 'version', render: (v: string) => <strong>{v}</strong> },
    { title: '框架', dataIndex: 'framework', key: 'framework', render: (f: string) => <Tag>{f || '-'}</Tag> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <Tag color={s === 'active' ? 'green' : 'default'}>{s}</Tag> },
    { title: '大小', dataIndex: 'size', key: 'size', render: (s: number) => s ? `${(s / 1024 / 1024).toFixed(2)} MB` : '-' },
    { title: '训练任务', dataIndex: 'training_job_name', key: 'training_job_name', render: (t: string) => t || '-' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at' },
    {
      title: '操作', key: 'actions',
      render: (_: unknown, record: ModelVersion) => (
        <Space size="small">
          <Button type="link" size="small" icon={<DownloadOutlined />} onClick={() => handleDownload(record.version)}>下载</Button>
          <Popconfirm title="确定删除此版本？" onConfirm={() => handleDeleteVersion(record.version)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/models')} style={{ marginBottom: 16 }}>
        返回列表
      </Button>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>模型详情: {model.name}</Title>
        <Space>
          <Button icon={<EditOutlined />} onClick={() => { editForm.setFieldsValue(model); setEditVisible(true); }}>编辑</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate(`/models/${name}/versions/create`)}>
            添加版本
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>
      </div>

      <Card style={{ marginBottom: 16 }}>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="名称">{model.name}</Descriptions.Item>
          <Descriptions.Item label="框架">{model.framework || '-'}</Descriptions.Item>
          <Descriptions.Item label="任务类型">{model.task_type || '-'}</Descriptions.Item>
          <Descriptions.Item label="负责人">{model.owner || '-'}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{model.description || '-'}</Descriptions.Item>
          <Descriptions.Item label="标签">{model.labels || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{model.created_at}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="版本列表">
        <Table columns={versionColumns} dataSource={versions} rowKey="id" pagination={false} />
      </Card>

      <Modal title="编辑模型" open={editVisible} onCancel={() => setEditVisible(false)} onOk={() => editForm.submit()}>
        <Form form={editForm} layout="vertical" onFinish={handleEdit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="labels" label="标签">
            <Input placeholder="逗号分隔的标签" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}