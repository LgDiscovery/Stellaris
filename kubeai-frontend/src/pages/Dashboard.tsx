import { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Typography, Spin, Tag, Space } from 'antd';
import {
  AppstoreOutlined,
  ThunderboltOutlined,
  UnorderedListOutlined,
  CheckCircleOutlined,
  ClusterOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { modelsApi } from '../api/models';
import { jobsApi } from '../api/jobs';
import { inferenceApi } from '../api/inference';

const { Title, Paragraph } = Typography;

interface Stats {
  totalModels: number;
  activeModels: number;
  totalTrainingJobs: number;
  succeededJobs: number;
  runningJobs: number;
  failedJobs: number;
  inferenceExecuted: number;
}

const serviceStatuses = [
  { name: 'API网关', url: '/auth/metrics', port: 8080, icon: <ApiOutlined /> },
  { name: '模型管理服务', url: '/models/health', port: 58080, icon: <AppstoreOutlined /> },
  { name: '任务调度服务', url: '/jobs/health', port: 58081, icon: <UnorderedListOutlined /> },
  { name: '推理网关服务', url: '/inference/health', port: 58082, icon: <ThunderboltOutlined /> },
];

export default function DashboardPage() {
  const [stats, setStats] = useState<Stats>({
    totalModels: 0,
    activeModels: 0,
    totalTrainingJobs: 0,
    succeededJobs: 0,
    runningJobs: 0,
    failedJobs: 0,
    inferenceExecuted: 0,
  });
  const [loading, setLoading] = useState(true);
  const [serviceHealth, setServiceHealth] = useState<Record<string, boolean>>({});

  useEffect(() => {
    async function fetchData() {
      try {
        const [modelsRes, trainJobsRes, inferenceJobsRes] = await Promise.allSettled([
          modelsApi.list({ page: 1, page_size: 100 }),
          jobsApi.listTasks({ task_type: 'training', page: 1, page_size: 100 }),
          jobsApi.listTasks({ task_type: 'inference', page: 1, page_size: 100 }),
        ]);

        let totalModels = 0;
        let activeModels = 0;
        if (modelsRes.status === 'fulfilled') {
          totalModels = modelsRes.value.data.total;
          activeModels = modelsRes.value.data.items.filter(
            (m) => m.versions?.some((v) => v.status === 'active')
          ).length;
        }

        let totalTrainingJobs = 0;
        let succeededJobs = 0;
        let runningJobs = 0;
        let failedJobs = 0;
        if (trainJobsRes.status === 'fulfilled') {
          totalTrainingJobs = trainJobsRes.value.data.total;
          trainJobsRes.value.data.items.forEach((item: any) => {
            if (item.status === 'succeeded') succeededJobs++;
            if (item.status === 'running') runningJobs++;
            if (item.status === 'failed') failedJobs++;
          });
        }

        let inferenceExecuted = 0;
        if (inferenceJobsRes.status === 'fulfilled') {
          inferenceExecuted = inferenceJobsRes.value.data.total;
        }

        setStats({ totalModels, activeModels, totalTrainingJobs, succeededJobs, runningJobs, failedJobs, inferenceExecuted });
      } catch {
        // 静默处理
      } finally {
        setLoading(false);
      }
    }

    async function checkHealth() {
      const checks: Record<string, boolean> = {};
      for (const svc of serviceStatuses) {
        try {
          const url = svc.url.startsWith('/auth') ? '/auth/metrics' : svc.url;
          await fetch(url);
          checks[svc.name] = true;
        } catch {
          checks[svc.name] = false;
        }
      }
      setServiceHealth(checks);
    }

    fetchData();
    checkHealth();
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" tip="加载仪表盘数据..." />
      </div>
    );
  }

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        <ClusterOutlined style={{ marginRight: 8 }} />
        平台概览
      </Title>
      <Paragraph type="secondary" style={{ marginBottom: 24 }}>
        KubeAI 云原生AI平台 · 模型管理 / 离线训练 / 在线推理 全流程闭环
      </Paragraph>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card" bordered={false} style={{ background: 'linear-gradient(135deg, #e6f4ff, #fff)' }}>
            <Statistic
              title="模型总数"
              value={stats.totalModels}
              prefix={<AppstoreOutlined style={{ color: '#1677ff' }} />}
              suffix={<Tag color="blue">已注册</Tag>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card" bordered={false} style={{ background: 'linear-gradient(135deg, #f6ffed, #fff)' }}>
            <Statistic
              title="活跃模型"
              value={stats.activeModels}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              suffix={<Tag color="green">Active</Tag>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card" bordered={false} style={{ background: 'linear-gradient(135deg, #fff7e6, #fff)' }}>
            <Statistic
              title="训练任务"
              value={stats.totalTrainingJobs}
              prefix={<UnorderedListOutlined style={{ color: '#fa8c16' }} />}
              suffix={
                <Space size={4}>
                  {stats.runningJobs > 0 && <Tag color="processing">{stats.runningJobs} 运行中</Tag>}
                  {stats.succeededJobs > 0 && <Tag color="success">{stats.succeededJobs} 成功</Tag>}
                </Space>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card" bordered={false} style={{ background: 'linear-gradient(135deg, #f9f0ff, #fff)' }}>
            <Statistic
              title="推理执行次数"
              value={stats.inferenceExecuted}
              prefix={<ThunderboltOutlined style={{ color: '#722ed1' }} />}
              suffix={<Tag color="purple">推理</Tag>}
            />
          </Card>
        </Col>
      </Row>

      <Card title="服务健康状态" bordered={false} style={{ borderRadius: 8 }}>
        <Row gutter={[16, 16]}>
          {serviceStatuses.map((svc) => (
            <Col xs={24} sm={12} md={6} key={svc.name}>
              <Card
                size="small"
                bordered
                style={{ textAlign: 'center', borderRadius: 6 }}
                bodyStyle={{ padding: '16px 12px' }}
              >
                <div style={{ fontSize: 28, marginBottom: 8 }}>{svc.icon}</div>
                <div style={{ fontWeight: 600, marginBottom: 4 }}>{svc.name}</div>
                <Tag color={serviceHealth[svc.name] ? 'success' : 'error'}>
                  {serviceHealth[svc.name] ? '健康' : '不可达'}
                </Tag>
                <div style={{ fontSize: 11, color: '#999', marginTop: 4 }}>端口: {svc.port}</div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} md={12}>
          <Card title="平台架构" bordered={false} style={{ borderRadius: 8 }}>
            <Paragraph style={{ fontSize: 13, lineHeight: 2 }}>
              <strong>接入层：</strong>统一API网关入口，JWT鉴权 + 流量限流
              <br />
              <strong>核心服务层：</strong>模型管理 / 推理服务网关 / 任务调度 三大微服务
              <br />
              <strong>编排层：</strong>K8s CRD + Operator 声明式管理
              <br />
              <strong>基础设施：</strong>PostgreSQL / MinIO / Redis / Prometheus
            </Paragraph>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="快速链接" bordered={false} style={{ borderRadius: 8 }}>
            <Paragraph style={{ fontSize: 13, lineHeight: 2.5 }}>
              📦 <strong>模型仓库：</strong>MinIO (端口 9000) - 模型文件存储
              <br />
              🗄️ <strong>元数据库：</strong>PostgreSQL (端口 5432) - 结构化数据
              <br />
              ⚡ <strong>缓存队列：</strong>Redis Cluster - 任务队列 & 分布式锁
              <br />
              📊 <strong>监控面板：</strong>Prometheus + Grafana - 全链路可观测
            </Paragraph>
          </Card>
        </Col>
      </Row>
    </div>
  );
}