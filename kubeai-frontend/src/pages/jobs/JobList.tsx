import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Space, Tag, Typography, Select, message, Popconfirm, Tabs } from 'antd';
import { ReloadOutlined, PlusOutlined, StopOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { jobsApi, InferenceTask, TrainingTask } from '../../api/jobs';

const { Title } = Typography;

export default function JobListPage() {
  const [activeTab, setActiveTab] = useState<'training' | 'inference'>('training');
  const [trainingTasks, setTrainingTasks] = useState<TrainingTask[]>([]);
  const [inferenceTasks, setInferenceTasks] = useState<InferenceTask[]>([]);
  const [totalTraining, setTotalTraining] = useState(0);
  const [totalInference, setTotalInference] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState<string | undefined>();
  const navigate = useNavigate();

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      if (activeTab === 'training') {
        const res = await jobsApi.listTasks({
          task_type: 'training',
          status: statusFilter,
          page,
          page_size: pageSize,
        });
        setTrainingTasks(res.data.items as TrainingTask[]);
        setTotalTraining(res.data.total);
      } else {
        const res = await jobsApi.listTasks({
          task_type: 'inference',
          status: statusFilter,
          page,
          page_size: pageSize,
        });
        setInferenceTasks(res.data.items as InferenceTask[]);
        setTotalInference(res.data.total);
      }
    } catch {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  }, [activeTab, page, pageSize, statusFilter]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleCancel = async (taskId: string) => {
    try {
      await jobsApi.callbackStatus({ task_id: taskId, status: 'cancelled', task_type: activeTab });
      message.success('任务已取消');
      fetchTasks();
    } catch {
      // 错误已处理
    }
  };

  const statusTag = (status: string) => {
    const map: Record<string, { color: string; text: string }> = {
      pending: { color: 'gold', text: '等待中' },
      running: { color: 'processing', text: '运行中' },
      succeeded: { color: 'success', text: '成功' },
      failed: { color: 'error', text: '失败' },
      cancelled: { color: 'default', text: '已取消' },
    };
    const info = map[status] || { color: 'default', text: status };
    return <Tag color={info.color}>{info.text}</Tag>;
  };

  const trainingColumns = [
    { title: '任务ID', dataIndex: 'task_id', key: 'task_id', width: 180, render: (t: string) => <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{t}</span> },
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '框架', dataIndex: 'framework', key: 'framework', render: (f: string) => <Tag>{f}</Tag> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => statusTag(s) },
    { title: '优先级', dataIndex: 'priority', key: 'priority' },
    { title: '重试', key: 'retry', render: (_: unknown, r: TrainingTask) => `${r.retry_count}/${r.max_retries}` },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at' },
    {
      title: '操作', key: 'actions',
      render: (_: unknown, record: TrainingTask) => (
        <Space size="small">
          {record.status === 'running' && (
            <Popconfirm title="确定取消此任务？" onConfirm={() => handleCancel(record.task_id)}>
              <Button type="link" size="small" danger icon={<StopOutlined />}>取消</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  const inferenceColumns = [
    { title: '任务ID', dataIndex: 'task_id', key: 'task_id', width: 180, render: (t: string) => <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{t}</span> },
    { title: '模型', dataIndex: 'model_name', key: 'model_name' },
    { title: '版本', dataIndex: 'model_version', key: 'model_version' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => statusTag(s) },
    { title: '优先级', dataIndex: 'priority', key: 'priority' },
    { title: 'Pod', dataIndex: 'pod_name', key: 'pod_name', render: (p: string) => p || '-' },
    { title: '创建时间', dataIndex: 'create_time', key: 'create_time' },
    {
      title: '操作', key: 'actions',
      render: (_: unknown, record: InferenceTask) => (
        <Space size="small">
          {record.status === 'running' && (
            <Popconfirm title="确定取消此任务？" onConfirm={() => handleCancel(record.task_id)}>
              <Button type="link" size="small" danger icon={<StopOutlined />}>取消</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>任务列表</Title>
        <Space>
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 140 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { label: '等待中', value: 'pending' },
              { label: '运行中', value: 'running' },
              { label: '成功', value: 'succeeded' },
              { label: '失败', value: 'failed' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={fetchTasks}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/jobs/training/create')}>
            提交训练任务
          </Button>
          <Button icon={<PlusOutlined />} onClick={() => navigate('/jobs/inference/create')}>
            提交推理任务
          </Button>
        </Space>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={(key) => {
          setActiveTab(key as 'training' | 'inference');
          setPage(1);
        }}
        items={[
          {
            key: 'training',
            label: `训练任务 (${totalTraining})`,
            children: (
              <Table
                columns={trainingColumns}
                dataSource={trainingTasks}
                rowKey="task_id"
                loading={loading}
                pagination={{
                  current: page,
                  pageSize,
                  total: totalTraining,
                  showTotal: (t) => `共 ${t} 个训练任务`,
                  onChange: (p, ps) => { setPage(p); setPageSize(ps); },
                }}
              />
            ),
          },
          {
            key: 'inference',
            label: `推理任务 (${totalInference})`,
            children: (
              <Table
                columns={inferenceColumns}
                dataSource={inferenceTasks}
                rowKey="task_id"
                loading={loading}
                pagination={{
                  current: page,
                  pageSize,
                  total: totalInference,
                  showTotal: (t) => `共 ${t} 个推理任务`,
                  onChange: (p, ps) => { setPage(p); setPageSize(ps); },
                }}
              />
            ),
          },
        ]}
      />
    </div>
  );
}