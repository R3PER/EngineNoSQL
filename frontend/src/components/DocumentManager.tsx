import React, { useState, useEffect, useRef } from 'react';
import {
  Card,
  Button,
  Input,
  Table,
  Modal,
  Form,
  Space,
  Typography,
  message,
  Popconfirm,
  Tag,
  Row,
  Col,
  Select,
  Divider,
  Switch,
  InputNumber,
  DatePicker,
  Checkbox,
  Radio,
  Tabs,
  App
} from 'antd';
import dayjs from 'dayjs';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  FileTextOutlined,
  UnorderedListOutlined,
  SaveOutlined,
  CloseOutlined
} from '@ant-design/icons';
import {
  QueryDocuments,
  InsertDocument,
  UpdateDocument,
  DeleteDocument,
  DeleteCollection,
  CreateIndex
} from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

// Component to render data in a formatted way
const DataRenderer: React.FC<{ data: Record<string, any> }> = ({ data }) => {
  const renderValue = (value: any, key: string): React.ReactNode => {
    if (value === null || value === undefined) {
      return <Text type="secondary">null</Text>;
    }
    
    if (typeof value === 'boolean') {
      return <Tag color={value ? 'green' : 'red'}>{value.toString()}</Tag>;
    }
    
    if (typeof value === 'number') {
      return <Text code style={{ color: '#1890ff' }}>{value}</Text>;
    }
    
    if (typeof value === 'string') {
      // Check if it's a date string
      if (value.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)) {
        return <Text type="secondary">{new Date(value).toLocaleString()}</Text>;
      }
      
      // Limit string length for display
      const maxLength = 30;
      if (value.length > maxLength) {
        return <Text>{value.substring(0, maxLength)}...</Text>;
      }
      return <Text>"{value}"</Text>;
    }
    
    if (Array.isArray(value)) {
      return (
        <div>
          <Text type="secondary">[Array({value.length})]</Text>
          {value.length > 0 && value.length <= 3 && (
            <div style={{ marginLeft: 8, fontSize: '12px' }}>
              {value.map((item, index) => (
                <div key={index}>• {renderValue(item, `${key}[${index}]`)}</div>
              ))}
            </div>
          )}
        </div>
      );
    }
    
    if (typeof value === 'object') {
      const keys = Object.keys(value);
      return (
        <div>
          <Text type="secondary">{'{'}{keys.length} fields{'}'}</Text>
          {keys.length <= 3 && (
            <div style={{ marginLeft: 8, fontSize: '12px' }}>
              {keys.map(k => (
                <div key={k}>
                  <Text strong>{k}:</Text> {renderValue(value[k], `${key}.${k}`)}
                </div>
              ))}
            </div>
          )}
        </div>
      );
    }
    
    return <Text>{String(value)}</Text>;
  };

  const dataKeys = Object.keys(data);
  
  return (
    <div style={{ maxWidth: 400 }}>
      {dataKeys.map(key => (
        <div key={key} style={{ marginBottom: 4 }}>
          <Text strong style={{ color: '#722ed1' }}>{key}:</Text>{' '}
          {renderValue(data[key], key)}
        </div>
      ))}
    </div>
  );
};

interface DocumentManagerProps {
  selectedDatabase: string | null;
  selectedCollection: string | null;
  collections: service.CollectionInfo[];
  onCollectionSelect: (collection: string) => void;
  sessionId: string | null;
  onDataChange?: () => Promise<void>;
  onCollectionDeleted?: (deletedCollectionName: string) => void;
}

const DocumentManager: React.FC<DocumentManagerProps> = ({
  selectedDatabase,
  selectedCollection,
  collections,
  onCollectionSelect,
  sessionId,
  onDataChange,
  onCollectionDeleted
}) => {
  const { message } = App.useApp();
  const [documents, setDocuments] = useState<service.DocumentResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingDoc, setEditingDoc] = useState<service.DocumentResponse | null>(null);
  const [searchField, setSearchField] = useState('');
  const [searchValue, setSearchValue] = useState('');
  const [indexField, setIndexField] = useState('');
  const [createMode, setCreateMode] = useState<'manual' | 'form'>('manual');
  const [formFields, setFormFields] = useState<Array<{id: string, name: string, type: string, value: any}>>([]);
  const [form] = Form.useForm();
  const [formBuilderForm] = Form.useForm();
  const [forceUpdateKey, setForceUpdateKey] = useState(0);
  const selectRef = useRef<any>(null);
  const [tempCollections, setTempCollections] = useState<service.CollectionInfo[]>([]);

  useEffect(() => {
    if (selectedDatabase && selectedCollection) {
      loadDocuments();
    }
  }, [selectedDatabase, selectedCollection]);

  // Clear selection if collection no longer exists - uruchamiany tylko przy rzeczywistych zmianach kolekcji
  useEffect(() => {
    if (selectedCollection) {
      // If collections is empty or undefined, clear selection
      if (!collections || collections.length === 0) {
        onCollectionSelect('');
        setDocuments([]);
        return;
      }
      
      // Check if selected collection still exists
      const collectionExists = collections.some(c => c.name === selectedCollection);
      if (!collectionExists) {
        onCollectionSelect('');
        setDocuments([]);
      }
    }
  }, [collections.length, selectedCollection, onCollectionSelect]);

  // Aktualizacja klucza wymuszającego re-render, ale tylko przy faktycznych zmianach
  useEffect(() => {
    setForceUpdateKey(prev => prev + 1);
  }, [collections.length]);

  const loadDocuments = async () => {
    if (!selectedDatabase || !selectedCollection || !sessionId) return;

    try {
      setLoading(true);
      const req = service.QueryRequest.createFrom({
        database: selectedDatabase,
        collection: selectedCollection,
        field: '',
        value: ''
      });
      const docs = await QueryDocuments(sessionId, req);
      setDocuments(docs);
    } catch (err) {
      message.error(`Failed to load documents: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    if (!selectedDatabase || !selectedCollection || !searchField || !searchValue || !sessionId) {
      await loadDocuments();
      return;
    }

    try {
      setLoading(true);
      const req = service.QueryRequest.createFrom({
        database: selectedDatabase,
        collection: selectedCollection,
        field: searchField,
        value: searchValue
      });
      const docs = await QueryDocuments(sessionId, req);
      setDocuments(docs);
    } catch (err) {
      message.error(`Failed to search documents: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDocument = () => {
    setEditingDoc(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEditDocument = (doc: service.DocumentResponse) => {
    setEditingDoc(doc);
    
    // Set form values for manual JSON mode
    form.setFieldsValue({
      id: doc.id,
      data: JSON.stringify(doc.data, null, 2)
    });

    // Set form builder form values
    formBuilderForm.setFieldsValue({
      id: doc.id
    });

    // Convert document data to form fields for Form Builder
    const fields = Object.keys(doc.data).map((key, index) => {
      const value = doc.data[key];
      let fieldType = 'string';
      let fieldValue = value;

      // Determine field type based on value
      if (typeof value === 'number') {
        fieldType = 'number';
      } else if (typeof value === 'boolean') {
        fieldType = 'boolean';
      } else if (Array.isArray(value)) {
        fieldType = 'array';
      } else if (typeof value === 'object' && value !== null) {
        fieldType = 'object';
      } else if (typeof value === 'string') {
        // Check if it's a date string
        if (value.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)) {
          fieldType = 'date';
          fieldValue = dayjs(value); // Convert to dayjs for DatePicker
        } else if (value.includes('@') && value.includes('.')) {
          fieldType = 'email';
        } else if (value.startsWith('http://') || value.startsWith('https://')) {
          fieldType = 'url';
        } else if (value.match(/^\+?[\d\s\-\(\)]+$/)) {
          fieldType = 'phone';
        }
      }

      return {
        id: `field_${Date.now()}_${index}`,
        name: key,
        type: fieldType,
        value: fieldValue
      };
    });

    setFormFields(fields);
    setModalVisible(true);
  };

  const handleDeleteDocument = async (docId: string) => {
    if (!selectedDatabase || !selectedCollection || !sessionId) return;

    try {
      const req = service.DeleteRequest.createFrom({
        database: selectedDatabase,
        collection: selectedCollection,
        id: docId
      });
      await DeleteDocument(sessionId, req);
      message.success('Document deleted successfully');
      await loadDocuments();
      // Refresh stats in DatabaseManager
      if (onDataChange && typeof onDataChange === 'function') {
        await onDataChange();
      }
    } catch (err) {
      message.error(`Failed to delete document: ${err}`);
    }
  };

  const handleModalOk = async () => {
    if (!sessionId) return;
    
    try {
      const values = await form.validateFields();
      const data = JSON.parse(values.data);

      if (editingDoc) {
        // Update existing document
        const req = service.UpdateRequest.createFrom({
          database: selectedDatabase!,
          collection: selectedCollection!,
          id: values.id,
          data: data
        });
        await UpdateDocument(sessionId, req);
        message.success('Document updated successfully');
      } else {
        // Create new document
        const req = service.InsertRequest.createFrom({
          database: selectedDatabase!,
          collection: selectedCollection!,
          id: values.id,
          data: data
        });
        await InsertDocument(sessionId, req);
        message.success('Document created successfully');
      }

      setModalVisible(false);
      await loadDocuments();
      // Refresh stats in DatabaseManager
      if (onDataChange && typeof onDataChange === 'function') {
        await onDataChange();
      }
    } catch (err) {
      if (err instanceof SyntaxError) {
        message.error('Invalid JSON format');
      } else {
        message.error(`Failed to save document: ${err}`);
      }
    }
  };

  const handleCreateIndex = async () => {
    if (!selectedDatabase || !selectedCollection || !indexField || !sessionId) {
      message.error('Please enter a field name for the index');
      return;
    }

    try {
      await CreateIndex(sessionId, selectedDatabase, selectedCollection, indexField);
      message.success('Index created successfully');
      setIndexField('');
    } catch (err) {
      message.error(`Failed to create index: ${err}`);
    }
  };

  const handleDeleteCollection = async () => {
    if (!selectedDatabase || !selectedCollection || !sessionId) return;

    const collectionToDelete = selectedCollection;
    
    try {
      setLoading(true);
      
      // Najpierw wyczyść selekcję i dokumenty natychmiast
      onCollectionSelect('');
      setDocuments([]);
      
      // Usuń kolekcję na backendzie
      await DeleteCollection(sessionId, selectedDatabase, collectionToDelete);
      message.success('Collection deleted successfully');
      
      // Natychmiast powiadom rodzica, aby usunął z stanu
      if (onCollectionDeleted) {
        onCollectionDeleted(collectionToDelete);
      }
      
      // Wymuś wyczyszczenie komponentu Select natychmiast
      if (selectRef.current) {
        selectRef.current.blur();
        selectRef.current.focus();
        selectRef.current.blur();
      }
      
      // Usuń bezpośrednio z lokalnego stanu tempCollections jeśli istnieje
      setTempCollections(prev => prev.filter(c => c.name !== collectionToDelete));
      
      // Wymuszenie ponownego renderowania, aby zaktualizować dropdown Select natychmiast
      setForceUpdateKey(prev => prev + 1);
      
      // Wymuś pełne odświeżenie danych
      if (onDataChange && typeof onDataChange === 'function') {
        try {
          await onDataChange();
        } catch (refreshErr) {
          console.error('Error refreshing data after collection deletion:', refreshErr);
        }
      }
      
      // Dodatkowe odświeżenie po krótkim opóźnieniu, aby upewnić się, że UI jest zaktualizowane
      setTimeout(() => {
        setForceUpdateKey(prev => prev + 1);
      }, 100);
      
    } catch (err) {
      message.error(`Failed to delete collection: ${err}`);
    } finally {
      setLoading(false);
    }
  };


  // Data types and predefined templates
  const dataTypes = [
    'string', 'number', 'boolean', 'date', 'array', 'object', 'email', 'url', 'phone', 'password'
  ];

  const predefinedTemplates = {
    'User Profile': [
      { name: 'firstName', type: 'string', value: '' },
      { name: 'lastName', type: 'string', value: '' },
      { name: 'email', type: 'email', value: '' },
      { name: 'age', type: 'number', value: 0 },
      { name: 'isActive', type: 'boolean', value: true },
      { name: 'createdAt', type: 'date', value: null },
    ],
    'Product': [
      { name: 'name', type: 'string', value: '' },
      { name: 'description', type: 'string', value: '' },
      { name: 'price', type: 'number', value: 0 },
      { name: 'inStock', type: 'boolean', value: true },
      { name: 'category', type: 'string', value: '' },
      { name: 'tags', type: 'array', value: [] },
    ],
    'Article/Blog Post': [
      { name: 'title', type: 'string', value: '' },
      { name: 'content', type: 'string', value: '' },
      { name: 'author', type: 'string', value: '' },
      { name: 'publishedAt', type: 'date', value: null },
      { name: 'isPublished', type: 'boolean', value: false },
      { name: 'tags', type: 'array', value: [] },
    ],
    'Company/Organization': [
      { name: 'companyName', type: 'string', value: '' },
      { name: 'industry', type: 'string', value: '' },
      { name: 'website', type: 'url', value: '' },
      { name: 'phone', type: 'phone', value: '' },
      { name: 'employees', type: 'number', value: 0 },
      { name: 'founded', type: 'date', value: null },
    ],
    'Order/Transaction': [
      { name: 'orderId', type: 'string', value: '' },
      { name: 'customerEmail', type: 'email', value: '' },
      { name: 'amount', type: 'number', value: 0 },
      { name: 'currency', type: 'string', value: 'USD' },
      { name: 'status', type: 'string', value: 'pending' },
      { name: 'orderDate', type: 'date', value: null },
    ],
    'Event': [
      { name: 'title', type: 'string', value: '' },
      { name: 'description', type: 'string', value: '' },
      { name: 'startDate', type: 'date', value: null },
      { name: 'endDate', type: 'date', value: null },
      { name: 'location', type: 'string', value: '' },
      { name: 'maxAttendees', type: 'number', value: 0 },
    ],
    'Custom': []
  };

  const addFormField = () => {
    const newField = {
      id: Date.now().toString(),
      name: '',
      type: 'string',
      value: ''
    };
    setFormFields([...formFields, newField]);
  };

  const removeFormField = (id: string) => {
    setFormFields(formFields.filter(field => field.id !== id));
  };

  const updateFormField = (id: string, updates: Partial<{name: string, type: string, value: any}>) => {
    setFormFields(formFields.map(field => 
      field.id === id ? { ...field, ...updates } : field
    ));
  };

  const loadTemplate = (templateName: string) => {
    const template = predefinedTemplates[templateName as keyof typeof predefinedTemplates];
    if (template) {
      setFormFields(template.map((field, index) => ({
        id: (Date.now() + index).toString(),
        ...field
      })));
    }
  };

  const handleFormBuilderOk = async () => {
    if (!sessionId) return;
    
    try {
      const values = await formBuilderForm.validateFields();
      
      // Build JSON object from form fields
      const data: Record<string, any> = {};
      formFields.forEach(field => {
        if (field.name) {
          // Convert dayjs dates to ISO strings for storage
          if (field.type === 'date' && field.value && dayjs.isDayjs(field.value)) {
            data[field.name] = field.value.toISOString();
          } else {
            data[field.name] = field.value;
          }
        }
      });

      if (editingDoc) {
        const req = service.UpdateRequest.createFrom({
          database: selectedDatabase!,
          collection: selectedCollection!,
          id: values.id,
          data: data
        });
        await UpdateDocument(sessionId, req);
        message.success('Document updated successfully');
      } else {
        const req = service.InsertRequest.createFrom({
          database: selectedDatabase!,
          collection: selectedCollection!,
          id: values.id,
          data: data
        });
        await InsertDocument(sessionId, req);
        message.success('Document created successfully');
      }

      setModalVisible(false);
      setFormFields([]);
      await loadDocuments();
    } catch (err) {
      message.error(`Failed to save document: ${err}`);
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      render: (text: string) => <Text code>{text}</Text>
    },
    {
      title: 'Data',
      dataIndex: 'data',
      key: 'data',
      render: (data: Record<string, any>) => <DataRenderer data={data} />
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (text: string) => <Text type="secondary">{text}</Text>
    },
    {
      title: 'Updated',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 150,
      render: (text: string) => <Text type="secondary">{text}</Text>
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: any, record: service.DocumentResponse) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEditDocument(record)}
          >
            Edit
          </Button>
          <Popconfirm
            title="Are you sure you want to delete this document?"
            onConfirm={() => handleDeleteDocument(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      )
    }
  ];

  if (!selectedDatabase) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <FileTextOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
          <Title level={4} type="secondary">Select a database to manage documents</Title>
        </div>
      </Card>
    );
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            <Title level={4} style={{ margin: 0 }}>Document Manager</Title>
          </Space>
        }
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={8}>
              <Select
                ref={selectRef}
                key={`collections-dropdown-${Math.random()}-${collections.length}-${forceUpdateKey}`}
                placeholder="Select collection"
                value={selectedCollection && collections.some(c => c.name === selectedCollection) ? selectedCollection : null}
                onChange={onCollectionSelect}
                style={{ width: '100%' }}
                allowClear
                options={collections.map(coll => ({
                  label: `${coll?.name} (${coll?.document_count || 0} docs)`,
                  value: coll?.name
                }))}
                notFoundContent={collections.length === 0 ? "No collections available" : "No matching collections"}
              />
          </Col>
          <Col xs={24} sm={16}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreateDocument}
                disabled={!selectedCollection}
              >
                Add Document
              </Button>
              {selectedCollection && (
                <Popconfirm
                  title="Are you sure you want to delete this collection?"
                  description="This action will permanently delete all documents in this collection!"
                  onConfirm={handleDeleteCollection}
                  okText="Yes, Delete"
                  cancelText="Cancel"
                  okType="danger"
                >
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                  >
                    Delete Collection
                  </Button>
                </Popconfirm>
              )}
            </Space>
          </Col>
        </Row>

        {selectedCollection && (
          <>
            <Divider />
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              <Col xs={24} md={12}>
                <Card size="small" title="Search Documents">
                  <Space.Compact style={{ width: '100%' }}>
                    <Input
                      placeholder="Field name"
                      value={searchField}
                      onChange={(e) => setSearchField(e.target.value)}
                      style={{ width: '40%' }}
                    />
                    <Input
                      placeholder="Search value"
                      value={searchValue}
                      onChange={(e) => setSearchValue(e.target.value)}
                      style={{ width: '40%' }}
                    />
                    <Button
                      type="primary"
                      icon={<SearchOutlined />}
                      onClick={handleSearch}
                      style={{ width: '20%' }}
                    >
                      Search
                    </Button>
                  </Space.Compact>
                </Card>
              </Col>
              <Col xs={24} md={12}>
                <Card size="small" title="Create Index">
                  <Space.Compact style={{ width: '100%' }}>
                    <Input
                      placeholder="Field name for index"
                      value={indexField}
                      onChange={(e) => setIndexField(e.target.value)}
                      style={{ width: '70%' }}
                    />
                    <Button
                      type="primary"
                      icon={<UnorderedListOutlined />}
                      onClick={handleCreateIndex}
                      style={{ width: '30%' }}
                    >
                      Create Index
                    </Button>
                  </Space.Compact>
                </Card>
              </Col>
            </Row>

            <Table
              columns={columns}
              dataSource={documents}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `Total ${total} documents`
              }}
              scroll={{ x: 800 }}
            />
          </>
        )}
      </Card>

      <Modal
        title={editingDoc ? 'Edit Document' : 'Create Document'}
        open={modalVisible}
        onOk={createMode === 'manual' ? handleModalOk : handleFormBuilderOk}
        onCancel={() => {
          setModalVisible(false);
          setFormFields([]);
          setCreateMode('manual');
        }}
        width={800}
        okText={editingDoc ? 'Update' : 'Create'}
      >
        <Tabs
          activeKey={createMode}
          onChange={(key) => setCreateMode(key as 'manual' | 'form')}
          items={[
            {
              key: 'manual',
              label: 'Manual JSON',
              children: (
                <Form form={form} layout="vertical">
                  <Form.Item
                    name="id"
                    label="Document ID"
                    rules={[{ required: true, message: 'Please enter document ID' }]}
                  >
                    <Input placeholder="Enter unique document ID" disabled={!!editingDoc} />
                  </Form.Item>
                  <Form.Item
                    name="data"
                    label="Document Data (JSON)"
                    rules={[
                      { required: true, message: 'Please enter document data' },
                      {
                        validator: (_, value) => {
                          try {
                            JSON.parse(value);
                            return Promise.resolve();
                          } catch {
                            return Promise.reject(new Error('Invalid JSON format'));
                          }
                        }
                      }
                    ]}
                  >
                    <TextArea
                      rows={12}
                      placeholder='{"name": "John", "age": 30, "city": "New York"}'
                    />
                  </Form.Item>
                </Form>
              )
            },
            {
              key: 'form',
              label: 'Form Builder',
              children: (
                <div>
                  <Form form={formBuilderForm} layout="vertical">
                    <Form.Item
                      name="id"
                      label="Document ID"
                      rules={[{ required: true, message: 'Please enter document ID' }]}
                    >
                      <Input placeholder="Enter unique document ID" disabled={!!editingDoc} />
                    </Form.Item>
                  </Form>
                  
                  <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
                    <div>
                      <Text strong>Template:</Text>
                      <Select
                        placeholder="Choose a template or start custom"
                        style={{ width: '100%', marginTop: 8 }}
                        onChange={loadTemplate}
                      >
                        {Object.keys(predefinedTemplates).map(template => (
                          <Option key={template} value={template}>{template}</Option>
                        ))}
                      </Select>
                    </div>
                  </Space>

                  <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                    {formFields.map((field) => (
                      <Card key={field.id} size="small" style={{ marginBottom: 8 }}>
                        <Row gutter={8} align="middle">
                          <Col span={6}>
                            <Input
                              placeholder="Field name"
                              value={field.name}
                              onChange={(e) => updateFormField(field.id, { name: e.target.value })}
                            />
                          </Col>
                          <Col span={4}>
                            <Select
                              value={field.type}
                              onChange={(value) => updateFormField(field.id, { type: value })}
                              style={{ width: '100%' }}
                            >
                              {dataTypes.map(type => (
                                <Option key={type} value={type}>{type}</Option>
                              ))}
                            </Select>
                          </Col>
                          <Col span={12}>
                            {field.type === 'string' && (
                              <Input
                                placeholder="Value"
                                value={field.value}
                                onChange={(e) => updateFormField(field.id, { value: e.target.value })}
                              />
                            )}
                            {field.type === 'number' && (
                              <InputNumber
                                placeholder="Value"
                                value={field.value}
                                onChange={(value) => updateFormField(field.id, { value })}
                                style={{ width: '100%' }}
                              />
                            )}
                            {field.type === 'boolean' && (
                              <Switch
                                checked={field.value}
                                onChange={(checked) => updateFormField(field.id, { value: checked })}
                              />
                            )}
                            {field.type === 'date' && (
                              <DatePicker
                                value={field.value}
                                onChange={(date) => updateFormField(field.id, { value: date })}
                                style={{ width: '100%' }}
                              />
                            )}
                            {field.type === 'email' && (
                              <Input
                                type="email"
                                placeholder="email@example.com"
                                value={field.value}
                                onChange={(e) => updateFormField(field.id, { value: e.target.value })}
                              />
                            )}
                            {field.type === 'url' && (
                              <Input
                                placeholder="https://example.com"
                                value={field.value}
                                onChange={(e) => updateFormField(field.id, { value: e.target.value })}
                              />
                            )}
                            {field.type === 'phone' && (
                              <Input
                                placeholder="+1234567890"
                                value={field.value}
                                onChange={(e) => updateFormField(field.id, { value: e.target.value })}
                              />
                            )}
                            {field.type === 'password' && (
                              <Input.Password
                                placeholder="Password"
                                value={field.value}
                                onChange={(e) => updateFormField(field.id, { value: e.target.value })}
                              />
                            )}
                            {field.type === 'array' && (
                              <Input
                                placeholder='["item1", "item2"]'
                                value={Array.isArray(field.value) ? JSON.stringify(field.value) : field.value}
                                onChange={(e) => {
                                  try {
                                    const parsed = JSON.parse(e.target.value);
                                    updateFormField(field.id, { value: parsed });
                                  } catch {
                                    updateFormField(field.id, { value: e.target.value });
                                  }
                                }}
                              />
                            )}
                            {field.type === 'object' && (
                              <TextArea
                                placeholder='{"key": "value"}'
                                value={typeof field.value === 'object' ? JSON.stringify(field.value) : field.value}
                                onChange={(e) => {
                                  try {
                                    const parsed = JSON.parse(e.target.value);
                                    updateFormField(field.id, { value: parsed });
                                  } catch {
                                    updateFormField(field.id, { value: e.target.value });
                                  }
                                }}
                                rows={2}
                              />
                            )}
                          </Col>
                          <Col span={2}>
                            <Button
                              type="link"
                              danger
                              icon={<DeleteOutlined />}
                              onClick={() => removeFormField(field.id)}
                            />
                          </Col>
                        </Row>
                      </Card>
                    ))}
                  </div>

                  <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={addFormField}
                    style={{ width: '100%', marginTop: 8 }}
                  >
                    Add Field
                  </Button>
                </div>
              )
            }
          ]}
        />
      </Modal>
    </Space>
  );
};

export default DocumentManager;
