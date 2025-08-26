import React, { useState } from 'react';
import { 
  Card, 
  Form, 
  Input, 
  Button, 
  Typography, 
  Space, 
  Alert,
  Tabs
} from 'antd';
import { 
  UserOutlined, 
  LockOutlined, 
  MailOutlined,
  DatabaseOutlined 
} from '@ant-design/icons';
import { Login, Register } from '../../wailsjs/go/main/App';

const { Title, Text } = Typography;

interface LoginFormProps {
  onLoginSuccess: (sessionId: string, user: any) => void;
}

interface LoginRequest {
  username: string;
  password: string;
}

interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

const LoginForm: React.FC<LoginFormProps> = ({ onLoginSuccess }) => {
  const [loginForm] = Form.useForm();
  const [registerForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('login');

  const handleLogin = async (values: LoginRequest) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await Login(values);
      
      if (response.success && response.session_id) {
        onLoginSuccess(response.session_id, response.user);
      } else {
        setError(response.message || 'Login failed');
      }
    } catch (err) {
      setError(`Login failed: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (values: RegisterRequest) => {
    try {
      setLoading(true);
      setError(null);
      setSuccess(null);
      
      const response = await Register(values);
      
      if (response.success) {
        setSuccess(response.message);
        registerForm.resetFields();
        setActiveTab('login');
      } else {
        setError(response.message);
      }
    } catch (err) {
      setError(`Registration failed: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ 
      minHeight: '100vh', 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '20px'
    }}>
      <Card
        style={{ 
          width: '100%', 
          maxWidth: 400,
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.3)'
        }}
      >
        <Space 
          direction="vertical" 
          size="large" 
          style={{ width: '100%', textAlign: 'center' }}
        >
          <div>
            <DatabaseOutlined style={{ fontSize: '48px', color: '#1890ff' }} />
            <Title level={2} style={{ margin: '10px 0 5px 0', color: '#1890ff' }}>
              EngineNoSQL
            </Title>
            <Text type="secondary">Professional NoSQL Database Management</Text>
          </div>

          {error && (
            <Alert
              message={error}
              type="error"
              showIcon
              closable
              onClose={() => setError(null)}
            />
          )}

          {success && (
            <Alert
              message={success}
              type="success"
              showIcon
              closable
              onClose={() => setSuccess(null)}
            />
          )}

          <Tabs 
            activeKey={activeTab} 
            onChange={setActiveTab}
            centered
            items={[
              {
                key: 'login',
                label: 'Login',
                children: (
                  <Form
                    form={loginForm}
                    name="login"
                    onFinish={handleLogin}
                    layout="vertical"
                    size="large"
                  >
                    <Form.Item
                      name="username"
                      rules={[
                        { required: true, message: 'Please enter your username!' }
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined />}
                        placeholder="Username"
                        autoComplete="username"
                      />
                    </Form.Item>

                    <Form.Item
                      name="password"
                      rules={[
                        { required: true, message: 'Please enter your password!' }
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Password"
                        autoComplete="current-password"
                      />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={loading}
                        style={{ width: '100%' }}
                      >
                        Sign In
                      </Button>
                    </Form.Item>
                  </Form>
                )
              },
              {
                key: 'register',
                label: 'Register',
                children: (
                  <Form
                    form={registerForm}
                    name="register"
                    onFinish={handleRegister}
                    layout="vertical"
                    size="large"
                  >
                    <Form.Item
                      name="username"
                      rules={[
                        { required: true, message: 'Please enter a username!' },
                        { min: 3, message: 'Username must be at least 3 characters!' }
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined />}
                        placeholder="Username"
                        autoComplete="username"
                      />
                    </Form.Item>

                    <Form.Item
                      name="email"
                      rules={[
                        { required: true, message: 'Please enter your email!' },
                        { type: 'email', message: 'Please enter a valid email!' }
                      ]}
                    >
                      <Input
                        prefix={<MailOutlined />}
                        placeholder="Email"
                        autoComplete="email"
                      />
                    </Form.Item>

                    <Form.Item
                      name="password"
                      rules={[
                        { required: true, message: 'Please enter a password!' },
                        { min: 6, message: 'Password must be at least 6 characters!' }
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Password"
                        autoComplete="new-password"
                      />
                    </Form.Item>

                    <Form.Item
                      name="confirmPassword"
                      dependencies={['password']}
                      rules={[
                        { required: true, message: 'Please confirm your password!' },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('password') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error('Passwords do not match!'));
                          },
                        }),
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Confirm Password"
                        autoComplete="new-password"
                      />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={loading}
                        style={{ width: '100%' }}
                      >
                        Sign Up
                      </Button>
                    </Form.Item>
                  </Form>
                )
              }
            ]}
          />

          <div style={{ textAlign: 'center', marginTop: '20px' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              Secure authentication with encrypted data storage
            </Text>
          </div>
        </Space>
      </Card>
    </div>
  );
};

export default LoginForm;
