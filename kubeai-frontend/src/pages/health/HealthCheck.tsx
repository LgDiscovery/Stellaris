import { useEffect, useState } from 'react';
import { Card, Row, Col, Tag, Typography, Spin, Button, Space, message, Divider } from 'antd';
import {
  HeartOutlined, CheckCircleOutlined, CloseCircleOutlined, ReloadOutlined,
  ApiOutlined, AppstoreOutlined, UnorderedListOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import { modelsApi } from '../../api/models';
import { inferenceApi } from '../../api/inference';
import { jobsApi } from '../../api/jobs';

const { Title, Text } = Typography;

interface ServiceStatus {
  name: string;
  icon: React.ReactNode;
  health: boolean | null;
  ready: boolean | null;
}

export default function HealthCheckPage() {
  const [services, setServices] = useState<ServiceStatus[]>([
    { name: 'API网关', icon: <ApiOutlined />, health: null, ready: null },
    { name: '模型管理服务', icon: <AppstoreOutlined />, health: null, ready: null },
    { name: '任务调度服务', icon: <UnorderedListOutlined />, health: null, ready: null },
    { name: '推理网关服务', icon: <ThunderboltOutlined />, health: null, ready: null },
  ]);
  const [checking, setChecking] = useState(false);

  const checkAll = async () => {
    setChecking(true);
    const updated = [...services];

    const checks: Array<{ index: number; healthFn: () => Promise<any>; readyFn: () => Promise<any> }> = [
      { index: 0, healthFn: () => fetch('/api/v1/auth/metrics'), readyFn: () => fetch('/api/v1/auth/metrics') },
      { index: 1, healthFn: () => modelsApi.healthCheck(), readyFn: () => modelsApi.readyCheck() },
      { index: 2, healthFn: () => jobsApi.healthCheck(), readyFn: () => jobsApi.readyCheck() },
      { index: 3, healthFn: () => inferenceApi.healthCheck(), readyFn: () => inferenceApi.readyCheck() },
    ];

    await Promise.all(
      checks.map(async ({ index, healthFn, readyFn }) => {
        try {
          await healthFn();
          updated[index].health = true;
        } catch {
          updated[index].health = false;
        }
        try {
          await readyFn();
          updated[index].ready = true;
        } catch {
          updated[index].ready = false;
        }
      })
    );

    setServices([...updated]);
    setChecking(false);
    message.info('健康检查完成');
  };

  useEffect(() => {
    checkAll();
  }, []);

  const allHealthy = services.every((s) => s.health === true && s.ready === true);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={4} style={{ margin: 0 }}>
          <HeartOutlined style={{ marginRight: 8, color: allHealthy ? '#52c41a' : '#ff4d4f' }} />
          健康检查
        </Title>
        <Button icon={<ReloadOutlined />} onClick={checkAll} loading={checking}>
          重新检查
        </Button>
      </div>

      <Card style={{ marginBottom: 16 }}>
        <div style={{ textAlign: 'center', padding: '12px 0' }}>
          <Text style={{ fontSize: 16 }}>
            系统整体状态：
          </Text>
          {checking ? (
            <Spin style={{ marginLeft: 12 }} />
          ) : allHealthy ? (
            <Tag color="success" style={{ fontSize: 14, marginLeft: 12 }}>
              <CheckCircleOutlined /> 全部健康
            </Tag>
          ) : (
            <Tag color="error" style={{ fontSize: 14, marginLeft: 12 }}>
              <CloseCircleOutlined /> 部分服务异常
            </Tag>
          )}
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        {services.map((svc) => (
          <Col xs={24} sm={12} md={6} key={svc.name}>
            <Card
              bordered={false}
              style={{
                textAlign: 'center',
                borderRadius: 8,
                boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
              }}
            >
              <div style={{ fontSize: 36, marginBottom: 12, color: '#1677ff' }}>{svc.icon}</div>
              <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 12 }}>{svc.name}</div>
              <Space direction="vertical" size={4}>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>Health: </Text>
                  {svc.health === null ? (
                    <Tag>检查中...</Tag>
                  ) : svc.health ? (
                    <Tag color="success">健康</Tag>
                  ) : (
                    <Tag color="error">异常</Tag>
                  )}
                </div>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>Ready: </Text>
                  {svc.ready === null ? (
                    <Tag>检查中...</Tag>
                  ) : svc.ready ? (
                    <Tag color="success">就绪</Tag>
                  ) : (
                    <Tag color="error">未就绪</Tag>
                  )}
                </div>
              </Space>
            </Card>
          </Col>
        ))}
      </Row>

      <Divider />
      <Card title="服务端点信息" size="small">
        <Row gutter={[16, 8]}>
          {[
            { name: 'API网关', url: '/api/v1/auth/metrics', port: '8080' },
            { name: '模型管理', url: '/api/v1/models/health', port: '58080' },
            { name: '任务调度', url: '/api/v1/jobs/health', port: '58081' },
            { name: '推理网关', url: '/api/v1/inference/health', port: '58082' },
          ].map((ep) => (
            <Col xs={24} sm={12} key={ep.name}>
              <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>
                {ep.name}: <Tag>{ep.url}</Tag> (端口 {ep.port})
              </Text>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );
}