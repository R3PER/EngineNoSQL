import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Input,
  Typography,
  Space,
  Row,
  Col,
  message,
  Popconfirm,
  Tag,
  Divider,
  Alert
} from 'antd';
import {
  KeyOutlined,
  CopyOutlined,
  ReloadOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  ApiOutlined
} from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;

interface ApiManagerProps {
  selectedDatabase: string | null;
  sessionId: string | null;
}

const ApiManager: React.FC<ApiManagerProps> = ({
  selectedDatabase,
  sessionId
}) => {
  const [apiKey, setApiKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [showKey, setShowKey] = useState(false);
  const [endpoint, setEndpoint] = useState('http://localhost:8080/api/v1');

  useEffect(() => {
    if (selectedDatabase && sessionId) {
      generateApiKey();
    }
  }, [selectedDatabase, sessionId]);

  const generateApiKey = () => {
    // Generate a random API key
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = 'ens_';
    for (let i = 0; i < 32; i++) {
      result += characters.charAt(Math.floor(Math.random() * characters.length));
    }
    setApiKey(result);
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('Copied to clipboard');
    } catch (err) {
      message.error('Failed to copy');
    }
  };

  const regenerateKey = () => {
    setLoading(true);
    setTimeout(() => {
      generateApiKey();
      setLoading(false);
      message.success('API key regenerated successfully');
    }, 1000);
  };

  if (!selectedDatabase) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <ApiOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
          <Title level={4} type="secondary">Select a database to manage API access</Title>
        </div>
      </Card>
    );
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card
        title={
          <Space>
            <ApiOutlined />
            <Title level={4} style={{ margin: 0 }}>API Management</Title>
          </Space>
        }
      >
        <Alert
          message="API Access"
          description="Use the API key below to access your database programmatically. Keep this key secure and do not share it publicly."
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card size="small" title="Database Information">
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text strong>Database:</Text>
                  <Tag color="blue" style={{ marginLeft: 8 }}>{selectedDatabase}</Tag>
                </div>
                <div>
                  <Text strong>API Endpoint:</Text>
                  <div style={{ marginTop: 8 }}>
                    <Space.Compact style={{ width: '100%' }}>
                      <Input
                        value={endpoint}
                        onChange={(e) => setEndpoint(e.target.value)}
                        style={{ width: 'calc(100% - 40px)' }}
                      />
                      <Button
                        icon={<CopyOutlined />}
                        onClick={() => copyToClipboard(endpoint)}
                      />
                    </Space.Compact>
                  </div>
                </div>
              </Space>
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card 
              size="small" 
              title="API Key"
              extra={
                <Popconfirm
                  title="Are you sure you want to regenerate the API key?"
                  description="This will invalidate the current key!"
                  onConfirm={regenerateKey}
                  okText="Yes, Regenerate"
                  cancelText="Cancel"
                >
                  <Button
                    icon={<ReloadOutlined />}
                    loading={loading}
                    size="small"
                  >
                    Regenerate
                  </Button>
                </Popconfirm>
              }
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text strong>Current API Key:</Text>
                  <div style={{ marginTop: 8 }}>
                    <Space.Compact style={{ width: '100%' }}>
                      <Input
                        value={showKey ? apiKey : '•'.repeat(apiKey.length)}
                        readOnly
                        style={{ width: 'calc(100% - 80px)' }}
                      />
                      <Button
                        icon={showKey ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                        onClick={() => setShowKey(!showKey)}
                      />
                      <Button
                        icon={<CopyOutlined />}
                        onClick={() => copyToClipboard(apiKey)}
                      />
                    </Space.Compact>
                  </div>
                </div>
              </Space>
            </Card>
          </Col>
        </Row>

        <Divider />

        <Card size="small" title="API Usage Examples">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <Text strong>Query Documents:</Text>
              <Paragraph copyable style={{ marginTop: 8, backgroundColor: '#f5f5f5', padding: 12, borderRadius: 4 }}>
{`curl -X POST "${endpoint}/query" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "database": "${selectedDatabase}",
    "collection": "your_collection",
    "field": "name",
    "value": "John"
  }'`}
              </Paragraph>
            </div>

            <div>
              <Text strong>Insert Document:</Text>
              <Paragraph copyable style={{ marginTop: 8, backgroundColor: '#f5f5f5', padding: 12, borderRadius: 4 }}>
{`curl -X POST "${endpoint}/insert" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "database": "${selectedDatabase}",
    "collection": "your_collection",
    "id": "doc123",
    "data": {"name": "John", "age": 30}
  }'`}
              </Paragraph>
            </div>

            <div>
              <Text strong>Update Document:</Text>
              <Paragraph copyable style={{ marginTop: 8, backgroundColor: '#f5f5f5', padding: 12, borderRadius: 4 }}>
{`curl -X PUT "${endpoint}/update" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "database": "${selectedDatabase}",
    "collection": "your_collection",
    "id": "doc123",
    "data": {"name": "John Doe", "age": 31}
  }'`}
              </Paragraph>
            </div>

            <div>
              <Text strong>Delete Document:</Text>
              <Paragraph copyable style={{ marginTop: 8, backgroundColor: '#f5f5f5', padding: 12, borderRadius: 4 }}>
{`curl -X DELETE "${endpoint}/delete" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "database": "${selectedDatabase}",
    "collection": "your_collection",
    "id": "doc123"
  }'`}
              </Paragraph>
            </div>
          </Space>
        </Card>
      </Card>
    </Space>
  );
};

export default ApiManager;
