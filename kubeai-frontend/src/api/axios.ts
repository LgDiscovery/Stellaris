import axios from 'axios';
import { message } from 'antd';

const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('kubeai_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;
    const msg = error.response?.data?.message || error.message || '请求失败';

    if (status === 401) {
      localStorage.removeItem('kubeai_token');
      localStorage.removeItem('kubeai_user');
      message.error('登录已过期，请重新登录');
      window.location.href = '/login';
    } else if (status === 403) {
      message.error('权限不足');
    } else if (status === 404) {
      message.error('资源不存在');
    } else if (status && status >= 500) {
      message.error('服务器内部错误');
    } else if (!error.response) {
      message.error('网络连接失败，请检查网络');
    } else {
      message.error(msg);
    }

    return Promise.reject(error);
  }
);

export default apiClient;