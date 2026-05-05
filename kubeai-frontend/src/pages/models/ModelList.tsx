import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Space, Tag, Typography, Input, Select, message, Popconfirm, Modal, Descriptions } from 'antd';
import { PlusOutlined, ReloadOutlined, EyeOutlined, DeleteOutlined, CloudUploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { modelsApi, Model } from '../../api/models';

const { Title } = Typography;

export default function ModelListPage() {
  const [models, setModels] = useState<Model[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [frameworkFilter, setFrameworkFilter] = useState<string | undefined>();
  const [searchText, setSearchText] = useState('');
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedModel, setSelectedModel] = useState<Model | null>(null);
  const navigate = useNavigate();

  const fetchModels = useCallback(async () => {
    setLoading(true);
    try {
      const res = await modelsApi.list({
        page,
        page_size: pageSize,
        framework: frameworkFilter || undefined,
      });
      setModels(res.data.items);
      setTotal(res.data.total);
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, frameworkFilter]);

  useEffect(() => {
    fetchModels();
  }, [fetchModels]);

  const handleDelete = async (name: string) => {
    try {
      await modelsApi.delete(name);
      message.success('模型已删除');
      fetchModels();
    } catch {
      // 错误已处理
    }
  };

  const handleViewDetail = async (name: string) => {
    try {
      const res = await modelsApi.getByName(name);
      setSelectedModel(res.data);
      setDetailVisible(true);
    } catch {
      // 错误已处理
    }
  };

  const filteredModels = models.filter((m) =>
    searchText ? m.name.toLowerCase().includes(searchText.toLowerCase()) : true
  );

  const columns = [
    {
      title: '模型名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <strong style={{ color: '#1677ff' }}>{name}</strong>,
    },
    {
      title: '框架',
      dataIndex: 'framework',
      key: 'framework',
      render: (f: string) => <Tag color="blue">{f || '未指定'}</Tag>,
    },
    {
      title: '任务类型',
      dataIndex: 'task_type',
      key: 'task_type',
      render: (t: string) => <Tag>{t || '通用'}</Tag>,
    },
    {
      title: '版本数',
      key: 'version_count',
      render: (_: unknown, record: Model) => record.versions?.length || 0,
    },
    {
      title: '负责人',
      dataIndex: 'owner',
      key: 'owner',
      render: (o: string) => o || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => t || '-',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: Model) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record.name)}
          >
            详情
          </Button>
          <Button
            type="link"
            size="small"
            icon={<CloudUploadOutlined />}
            onClick={() => navigate(`/models/${record.name}/versions/create`)}
          >
            添加版本
          </Button>
          <Popconfirm
            title="确定删除此模型？"
            onConfirm={() => handleDelete(record.name)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>
          模型列表
        </Title>
        <Space>
          <Input.Search
            placeholder="搜索模型名称"
            allowClear
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 220 }}
          />
          <Select
            placeholder="框架筛选"
            allowClear
            style={{ width: 140 }}
            value={frameworkFilter}
            onChange={setFrameworkFilter}
            options={[
              { label: 'PyTorch', value: 'PyTorch' },
              { label: 'TensorFlow', value: 'TensorFlow' },
              { label: 'ONNX', value: 'ONNX' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={fetchModels}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/models/create')}>
            创建模型
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={filteredModels}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 个模型`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />

      <Modal
        title="模型详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedModel && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="模型名称">{selectedModel.name}</Descriptions.Item>
            <Descriptions.Item label="框架">{selectedModel.framework || '-'}</Descriptions.Item>
            <Descriptions.Item label="任务类型">{selectedModel.task_type || '-'}</Descriptions.Item>
            <Descriptions.Item label="负责人">{selectedModel.owner || '-'}</Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>
              {selectedModel.description || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="标签">{selectedModel.labels || '-'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{selectedModel.created_at}</Descriptions.Item>
            <Descriptions.Item label="版本列表" span={2}>
              {selectedModel.versions?.length ? (
                <Space wrap>
                  {selectedModel.versions.map((v) => (
                    <Tag key={v.id} color={v.status === 'active' ? 'green' : 'default'}>
                      {v.version} ({v.status})
                    </Tag>
                  ))}
                </Space>
              ) : (
                '暂无版本'
              )}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}