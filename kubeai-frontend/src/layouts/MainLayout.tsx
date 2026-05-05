import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, Avatar, Dropdown, theme, Typography } from 'antd';
import {
  DashboardOutlined,
  AppstoreOutlined,
  ThunderboltOutlined,
  UnorderedListOutlined,
  HeartOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  UserOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
  {
    key: '/models',
    icon: <AppstoreOutlined />,
    label: '模型管理',
    children: [
      { key: '/models', label: '模型列表' },
      { key: '/models/create', label: '创建模型' },
    ],
  },
  {
    key: '/inference',
    icon: <ThunderboltOutlined />,
    label: '推理服务',
    children: [
      { key: '/inference/execute', label: '执行推理' },
    ],
  },
  {
    key: '/jobs',
    icon: <UnorderedListOutlined />,
    label: '任务调度',
    children: [
      { key: '/jobs', label: '任务列表' },
      { key: '/jobs/training/create', label: '提交训练任务' },
      { key: '/jobs/inference/create', label: '提交推理任务' },
    ],
  },
  { key: '/health', icon: <HeartOutlined />, label: '健康检查' },
];

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { token: themeToken } = theme.useToken();

  const handleMenuClick = (info: { key: string }) => {
    navigate(info.key);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人信息' },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true, onClick: handleLogout },
  ];

  const getSelectedKeys = () => {
    const path = location.pathname;
    if (path.startsWith('/models')) return ['/models'];
    if (path.startsWith('/inference')) return ['/inference'];
    if (path.startsWith('/jobs')) return ['/jobs'];
    if (path.startsWith('/health')) return ['/health'];
    return [path];
  };

  const getOpenKeys = () => {
    const path = location.pathname;
    if (path.startsWith('/models')) return ['/models'];
    if (path.startsWith('/inference')) return ['/inference'];
    if (path.startsWith('/jobs')) return ['/jobs'];
    return [];
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        style={{
          background: 'linear-gradient(180deg, #001529 0%, #002140 100%)',
          borderRight: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid rgba(255,255,255,0.08)',
          }}
        >
          <ApiOutlined style={{ fontSize: collapsed ? 24 : 28, color: '#69b1ff', marginRight: collapsed ? 0 : 10 }} />
          {!collapsed && <span className="logo-text">Kube<span>AI</span></span>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ background: 'transparent', borderRight: 'none', marginTop: 8 }}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
            boxShadow: '0 1px 4px rgba(0,0,0,0.04)',
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: 16, width: 40, height: 40 }}
          />
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar
                style={{ backgroundColor: themeToken.colorPrimary }}
                icon={<UserOutlined />}
                size="small"
              />
              <Text strong>{user?.username || '用户'}</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {user?.role === 'admin' ? '管理员' : '用户'}
              </Text>
            </div>
          </Dropdown>
        </Header>
        <Content
          style={{
            margin: 16,
            padding: 24,
            background: '#fff',
            borderRadius: 8,
            minHeight: 280,
            overflow: 'auto',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}