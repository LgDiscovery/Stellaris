import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi, LoginRequest, LoginResponse } from '../api/auth';

interface AuthUser {
  userId: number;
  username: string;
  role: string;
}

interface AuthContextType {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (data: LoginRequest) => Promise<LoginResponse>;
  register: (data: LoginRequest) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const savedToken = localStorage.getItem('kubeai_token');
    const savedUser = localStorage.getItem('kubeai_user');
    if (savedToken && savedUser) {
      try {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
      } catch {
        localStorage.removeItem('kubeai_token');
        localStorage.removeItem('kubeai_user');
      }
    }
    setLoading(false);
  }, []);

  const login = useCallback(async (data: LoginRequest) => {
    const response = await authApi.login(data);
    const { token: newToken, user_id, username, role } = response.data;
    const userData: AuthUser = { userId: user_id, username, role };
    localStorage.setItem('kubeai_token', newToken);
    localStorage.setItem('kubeai_user', JSON.stringify(userData));
    setToken(newToken);
    setUser(userData);
    return response.data;
  }, []);

  const register = useCallback(async (data: LoginRequest) => {
    await authApi.register(data);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('kubeai_token');
    localStorage.removeItem('kubeai_user');
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token && !!user,
        loading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}