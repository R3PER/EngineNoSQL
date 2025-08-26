import React, { useState, useEffect, useCallback } from 'react';
import { 
  Card, 
  Button, 
  Input, 
  List, 
  Typography, 
  Space, 
  Statistic, 
  Row, 
  Col, 
  Form, 
  App,
  Spin,
  Badge,
  Divider,
  Popconfirm,
  Modal,
  Upload,
  Select,
  Radio
} from 'antd';
import { 
  DatabaseOutlined, 
  FolderOutlined, 
  PlusOutlined, 
  FileTextOutlined,
  BarChartOutlined,
  DeleteOutlined,
  ExportOutlined,
  ImportOutlined,
  DownloadOutlined,
  UploadOutlined
} from '@ant-design/icons';
import { 
  CreateDatabase, 
  DeleteDatabase,
  ListDatabases, 
  CreateCollection, 
  GetCollections,
  GetDatabaseStats,
  ExportData,
  ImportData,
  ImportDataFromContent,
  CreateBackup,
  RestoreBackup,
  ListBackups,
  GetSupportedFormats
} from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';

const { Title, Text } = Typography;

interface DatabaseManagerProps {
  onDatabaseSelect: (database: string) => void;
  selectedDatabase: string | null;
  sessionId: string | null;
  onCollectionsChange?: (collections: service.CollectionInfo[]) => void;
  onRefreshNeeded?: (refreshFn: () => Promise<void>) => void;
  onStatsUpdateNeeded?: (updateStatsFn: () => void) => void;
  onCollectionRemovalNeeded?: (removeCollectionFn: (collectionName: string) => void) => void;
}

const DatabaseManager: React.FC<DatabaseManagerProps> = ({ 
  onDatabaseSelect, 
  selectedDatabase,
  sessionId,
  onCollectionsChange,
  onRefreshNeeded,
  onStatsUpdateNeeded,
  onCollectionRemovalNeeded
}) => {

  const { message } = App.useApp();
  const [databases, setDatabases] = useState<service.DatabaseInfo[]>([]);
  const [collections, setCollections] = useState<service.CollectionInfo[]>([]);
  const [newDbName, setNewDbName] = useState('');
  const [newCollName, setNewCollName] = useState('');
  const [stats, setStats] = useState<Record<string, any> | null>(null);

  // Expose refresh functions to parent
  const refreshData = useCallback(async () => {
    if (selectedDatabase && sessionId) {
      await loadCollections();
      await loadStats();
    }
  }, [selectedDatabase, sessionId]);

  // Function to immediately update stats when collection is deleted
  const updateStatsAfterCollectionDelete = useCallback(() => {
    if (stats) {
      setStats(prevStats => ({
        ...prevStats,
        collections_count: Math.max(0, (prevStats?.collections_count || 0) - 1)
      }));
    }
  }, [stats]);

  // Funkcja do natychmiastowego usunięcia kolekcji z lokalnej listy
  const removeCollectionFromList = useCallback((collectionName: string) => {
    // Usuń kolekcję z lokalnego stanu
    setCollections(prevCollections => {
      const newCollections = prevCollections.filter(col => col.name !== collectionName);
      return newCollections;
    });
    
    // Natychmiast po usunięciu kolekcji, zaktualizuj statystyki
    if (stats) {
      // Znajdź usuniętą kolekcję, aby odjąć właściwą liczbę dokumentów
      const deletedCollection = collections.find(c => c.name === collectionName);
      const docsToRemove = deletedCollection?.document_count || 0;
      
      setStats(prevStats => ({
        ...prevStats,
        collections_count: Math.max(0, (prevStats?.collections_count || 1) - 1),
        total_documents: Math.max(0, (prevStats?.total_documents || 0) - docsToRemove)
      }));
    }
    
    // Odśwież listę kolekcji z backendu bez zbędnych logów
    if (selectedDatabase && sessionId) {
      loadCollections().catch(() => {
        // Cicha obsługa błędów - unikamy spamu w konsoli
      });
    }
  }, [collections, selectedDatabase, sessionId, stats]);

  // Pass refresh function to parent
  useEffect(() => {
    if (onRefreshNeeded) {
      onRefreshNeeded(refreshData);
    }
  }, [onRefreshNeeded, refreshData]);

  // Pass stats update function to parent
  useEffect(() => {
    if (onStatsUpdateNeeded) {
      onStatsUpdateNeeded(updateStatsAfterCollectionDelete);
    }
  }, [onStatsUpdateNeeded, updateStatsAfterCollectionDelete]);

  // Pass collection removal function to parent
  useEffect(() => {
    if (onCollectionRemovalNeeded) {
      onCollectionRemovalNeeded(removeCollectionFromList);
    }
  }, [onCollectionRemovalNeeded, removeCollectionFromList]);

  // Notify parent about collections changes - tylko przy faktycznych zmianach
  useEffect(() => {
    if (onCollectionsChange && collections.length > 0) {
      onCollectionsChange(collections);
    }
  }, [collections.length, onCollectionsChange]);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [exportModalVisible, setExportModalVisible] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [exportPath, setExportPath] = useState('');
  const [exportFormat, setExportFormat] = useState('json');
  const [exportName, setExportName] = useState('');
  const [exportType, setExportType] = useState<'database' | 'collection'>('database');
  const [selectedExportCollection, setSelectedExportCollection] = useState<string>('');
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importCollectionName, setImportCollectionName] = useState('');
  const [createCollection, setCreateCollection] = useState(true);
  const [overwriteData, setOverwriteData] = useState(false);

  useEffect(() => {
    loadDatabases();
  }, [sessionId]);

  useEffect(() => {
    if (selectedDatabase) {
      loadCollections();
      loadStats();
    }
  }, [selectedDatabase]);

  const loadDatabases = async () => {
    if (!sessionId) return;
    
    try {
      setLoading(true);
      const dbList = await ListDatabases(sessionId);
      setDatabases(dbList || []);
      setError(null);
    } catch (err) {
      setError(`Failed to load databases: ${err}`);
      setDatabases([]);
    } finally {
      setLoading(false);
    }
  };

  const loadCollections = async () => {
    if (!selectedDatabase || !sessionId) return;
    
    try {
      const collList = await GetCollections(sessionId, selectedDatabase);
      // Handle case where GetCollections returns null
      const collections = collList || [];
      setCollections(collections);
    } catch (err) {
      setError(`Failed to load collections: ${err}`);
      // Set empty array on error
      const emptyCollections: service.CollectionInfo[] = [];
      setCollections(emptyCollections);
    }
  };

  const loadStats = async () => {
    if (!selectedDatabase || !sessionId) return;
    
    try {
      const dbStats = await GetDatabaseStats(sessionId, selectedDatabase);
      setStats(dbStats);
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  };

  const handleCreateDatabase = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newDbName.trim() || !sessionId) return;

    try {
      setLoading(true);
      await CreateDatabase(sessionId, newDbName.trim());
      setNewDbName('');
      await loadDatabases();
      setError(null);
    } catch (err) {
      setError(`Failed to create database: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateCollection = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCollName.trim() || !selectedDatabase || !sessionId) return;

    try {
      setLoading(true);
      await CreateCollection(sessionId, selectedDatabase, newCollName.trim());
      setNewCollName('');
      await loadCollections();
      setError(null);
    } catch (err) {
      setError(`Failed to create collection: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDatabaseSubmit = async () => {
    console.log('Database creation attempt:', { newDbName, sessionId });
    
    if (!newDbName.trim()) {
      message.error('Database name cannot be empty');
      return;
    }
    
    if (!sessionId) {
      message.error('No session found. Please login again.');
      return;
    }

    try {
      setLoading(true);
      console.log('Creating database:', newDbName.trim());
      await CreateDatabase(sessionId, newDbName.trim());
      setNewDbName('');
      await loadDatabases();
      message.success('Database created successfully');
    } catch (err) {
      console.error('Database creation error:', err);
      message.error(`Failed to create database: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateCollectionSubmit = async () => {
    if (!newCollName.trim() || !selectedDatabase || !sessionId) {
      message.error('Collection name cannot be empty');
      return;
    }

    try {
      setLoading(true);
      await CreateCollection(sessionId, selectedDatabase, newCollName.trim());
      setNewCollName('');
      await loadCollections();
      await loadStats(); // Refresh stats
      message.success('Collection created successfully');
    } catch (err) {
      message.error(`Failed to create collection: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteDatabase = async (dbName: string) => {
    if (!sessionId) return;

    try {
      setLoading(true);
      await DeleteDatabase(sessionId, dbName);
      message.success(`Database ${dbName} deleted successfully`);
      await loadDatabases();
      if (selectedDatabase === dbName) {
        onDatabaseSelect('');
      }
    } catch (err) {
      message.error(`Failed to delete database: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleExportDatabase = () => {
    setExportModalVisible(true);
    setExportPath('/home/');
    setExportName(`${selectedDatabase}_export`);
  };

  const handleExportModalOk = async () => {
    if (!selectedDatabase || !sessionId || !exportPath.trim() || !exportName.trim()) {
      message.error('Please enter export path and name prefix');
      return;
    }

    if (exportType === 'collection' && !selectedExportCollection) {
      message.error('Please select a collection to export');
      return;
    }

    try {
      setLoading(true);
      
      let collectionsToExport: service.CollectionInfo[] = [];
      
      if (exportType === 'database') {
        // Export entire database
        collectionsToExport = await GetCollections(sessionId, selectedDatabase);
      } else {
        // Export specific collection
        const allCollections = await GetCollections(sessionId, selectedDatabase);
        const selectedColl = allCollections.find(c => c.name === selectedExportCollection);
        if (selectedColl) {
          collectionsToExport = [selectedColl];
        }
      }
      
      if (collectionsToExport.length === 0) {
        message.warning('No collections to export');
        return;
      }

      let exportedCount = 0;
      for (const coll of collectionsToExport) {
        const fileName = exportType === 'database' ? 
          `${exportName}_${coll.name}.${exportFormat}` : 
          `${exportName}.${exportFormat}`;
        const fullPath = exportPath.endsWith('/') ? 
          `${exportPath}${fileName}` : 
          `${exportPath}/${fileName}`;
          
        const exportRequest = service.ExportRequest.createFrom({
          database: selectedDatabase,
          collection: coll.name,
          format: exportFormat,
          file_path: fullPath
        });
        
        await ExportData(sessionId, exportRequest);
        exportedCount++;
      }
      
      setExportModalVisible(false);
      setExportPath('');
      setExportName('');
      setSelectedExportCollection('');
      
      const exportTypeText = exportType === 'database' ? 'Database' : 'Collection';
      message.success(`${exportTypeText} exported successfully! ${exportedCount} ${exportedCount === 1 ? 'file' : 'files'} exported to ${exportPath}`);
    } catch (err) {
      message.error(`Failed to export ${exportType}: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleImportDatabase = async (file: File) => {
    if (!selectedDatabase || !sessionId) return;

    try {
      setLoading(true);
      
      // Read file content
      const content = await file.text();
      
      // Extract collection name from filename (remove extension)
      const fileName = file.name.replace(/\.(json|csv)$/, '');
      const collectionName = fileName.includes('_') ? 
        fileName.split('_').pop() || fileName : 
        fileName;
      
      // Determine format
      const format = file.name.endsWith('.csv') ? 'csv' : 'json';
      
      // Use the new ImportDataFromContent API
      const result = await ImportDataFromContent(
        sessionId,
        selectedDatabase,
        collectionName,
        content,
        format,
        true // createCollection
      );
      
      // Refresh all data after import
      await loadCollections();
      await loadStats();
      await loadDatabases();
      
      if (result.errors && result.errors.length > 0) {
        message.warning(`Import completed with warnings! ${result.imported} documents imported, ${result.skipped} skipped. Errors: ${result.errors.slice(0, 3).join(', ')}`);
      } else {
        message.success(`Import completed! ${result.imported} documents imported to collection "${collectionName}".`);
      }
    } catch (err) {
      message.error(`Failed to import database: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateBackup = async () => {
    if (!selectedDatabase || !sessionId) return;

    try {
      setLoading(true);
      const backupRequest = service.BackupRequest.createFrom({
        database: selectedDatabase
      });
      
      await CreateBackup(sessionId, backupRequest);
      message.success('Backup created successfully');
    } catch (err) {
      message.error(`Failed to create backup: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card 
        title={
          <Space>
            <DatabaseOutlined />
            <Title level={4} style={{ margin: 0 }}>Databases</Title>
          </Space>
        }
        extra={
          <Space>
            <Input
              placeholder="Database name"
              value={newDbName}
              onChange={(e) => setNewDbName(e.target.value)}
              onPressEnter={handleCreateDatabaseSubmit}
              style={{ width: 200 }}
            />
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleCreateDatabaseSubmit}
              loading={loading}
              disabled={!newDbName.trim()}
            >
              Create DB
            </Button>
          </Space>
        }
      >
        <Spin spinning={loading}>
          <Row gutter={[16, 16]}>
            {databases.map((db) => (
              <Col xs={24} sm={12} md={8} lg={6} key={db.name}>
                <Card
                  hoverable
                  onClick={() => onDatabaseSelect(db.name)}
                  style={{
                    cursor: 'pointer',
                    border: selectedDatabase === db.name ? '2px solid #1890ff' : '1px solid #d9d9d9',
                    boxShadow: selectedDatabase === db.name ? '0 4px 12px rgba(24, 144, 255, 0.15)' : undefined
                  }}
                  styles={{ body: { padding: '16px' } }}
                >
                  <Space direction="vertical" align="center" style={{ width: '100%' }}>
                    <DatabaseOutlined 
                      style={{ 
                        fontSize: '32px', 
                        color: selectedDatabase === db.name ? '#1890ff' : '#666'
                      }} 
                    />
                    <Title level={5} style={{ margin: '8px 0 4px 0', textAlign: 'center' }}>
                      {db.name}
                    </Title>
                    <Badge 
                      count={db.collections?.length || 0} 
                      showZero 
                      color="blue"
                      style={{ fontSize: '12px' }}
                    />
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {db.collections?.length || 0} collections
                    </Text>
                  </Space>
                </Card>
              </Col>
            ))}
            {databases.length === 0 && !loading && (
              <Col span={24}>
                <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                  <DatabaseOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                  <div>No databases found. Create your first database!</div>
                </div>
              </Col>
            )}
          </Row>
        </Spin>
      </Card>

      {selectedDatabase && (
        <>
          <Card 
            title={
              <Space>
                <DatabaseOutlined />
                <Title level={4} style={{ margin: 0 }}>Database Operations</Title>
              </Space>
            }
          >
            <Space wrap>
              <Button 
                type="primary" 
                icon={<ExportOutlined />}
                onClick={handleExportDatabase}
                loading={loading}
              >
                Export Database
              </Button>
              <Upload
                accept=".json"
                showUploadList={false}
                beforeUpload={(file) => {
                  handleImportDatabase(file);
                  return false;
                }}
              >
                <Button 
                  icon={<ImportOutlined />}
                  loading={loading}
                >
                  Import Database
                </Button>
              </Upload>
              <Button 
                icon={<DownloadOutlined />}
                onClick={handleCreateBackup}
                loading={loading}
              >
                Create Backup
              </Button>
              <Popconfirm
                title="Are you sure you want to delete this database?"
                description="This action cannot be undone!"
                onConfirm={() => handleDeleteDatabase(selectedDatabase)}
                okText="Yes, Delete"
                cancelText="Cancel"
                okType="danger"
              >
                <Button 
                  danger 
                  icon={<DeleteOutlined />}
                  loading={loading}
                >
                  Delete Database
                </Button>
              </Popconfirm>
            </Space>
          </Card>
          
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={16}>
              <Card 
                key={`collections-card-${collections.length}-${JSON.stringify(collections.map(c => c.name).sort())}`}
                title={
                  <Space>
                    <FolderOutlined />
                    <Title level={4} style={{ margin: 0 }}>
                      Collections in {selectedDatabase}
                    </Title>
                  </Space>
                }
                extra={
                  <Space>
                    <Input
                      placeholder="Collection name"
                      value={newCollName}
                      onChange={(e) => setNewCollName(e.target.value)}
                      onPressEnter={handleCreateCollectionSubmit}
                      style={{ width: 200 }}
                    />
                    <Button 
                      type="primary" 
                      icon={<PlusOutlined />}
                      onClick={handleCreateCollectionSubmit}
                      loading={loading}
                      disabled={!newCollName.trim()}
                    >
                      Create Collection
                    </Button>
                  </Space>
                }
              >
                <List
                  key={`collections-list-${collections.length}-${JSON.stringify(collections.map(c => c.name).sort())}`}
                  dataSource={collections || []}
                  locale={{
                    emptyText: collections.length === 0 ? "No collections found. Create your first collection!" : "No collections match your criteria."
                  }}
                  renderItem={(coll) => (
                    <List.Item key={`${coll.name}-${coll.document_count || 0}`}>
                      <List.Item.Meta
                        avatar={<FileTextOutlined style={{ fontSize: '18px', color: '#52c41a' }} />}
                        title={<Text strong>{coll?.name || 'Unknown'}</Text>}
                        description={
                          <Space split={<Divider type="vertical" />}>
                            <Badge count={coll?.document_count || 0} showZero color="green" text="documents" />
                            <Badge count={coll?.indexes?.length || 0} showZero color="orange" text="indexes" />
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>

            <Col xs={24} lg={8}>
              {stats && (
                <Card 
                  title={
                    <Space>
                      <BarChartOutlined />
                      <Title level={4} style={{ margin: 0 }}>Statistics</Title>
                    </Space>
                  }
                >
                  <Row gutter={[16, 16]}>
                    <Col span={24}>
                      <Statistic
                        title="Total Documents"
                        value={stats.total_documents}
                        valueStyle={{ color: '#3f8600' }}
                      />
                    </Col>
                    <Col span={24}>
                      <Statistic
                        title="Total Indexes"
                        value={stats.total_indexes}
                        valueStyle={{ color: '#cf1322' }}
                      />
                    </Col>
                    <Col span={24}>
                      <Statistic
                        title="Collections"
                        value={stats.collections_count}
                        valueStyle={{ color: '#1890ff' }}
                      />
                    </Col>
                  </Row>
                </Card>
              )}
            </Col>
          </Row>
        </>
      )}

      {/* Export Database Modal */}
      <Modal
        title={`Export Database: ${selectedDatabase}`}
        open={exportModalVisible}
        onOk={handleExportModalOk}
        onCancel={() => {
          setExportModalVisible(false);
          setExportPath('');
        }}
        width={600}
        confirmLoading={loading}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Text strong>Export Type:</Text>
            <Radio.Group
              value={exportType}
              onChange={(e) => setExportType(e.target.value)}
              style={{ marginTop: 8, display: 'block' }}
            >
              <Radio value="database">
                <Space>
                  <DatabaseOutlined />
                  Entire Database
                </Space>
              </Radio>
              <Radio value="collection">
                <Space>
                  <FolderOutlined />
                  Single Collection
                </Space>
              </Radio>
            </Radio.Group>
          </div>

          {exportType === 'collection' && (
            <div>
              <Text strong>Select Collection:</Text>
              <Select
                placeholder="Choose collection to export"
                value={selectedExportCollection}
                onChange={setSelectedExportCollection}
                style={{ width: '100%', marginTop: 8 }}
              >
                {collections.map(coll => (
                  <Select.Option key={coll.name} value={coll.name}>
                    {coll.name} ({coll.document_count} docs)
                  </Select.Option>
                ))}
              </Select>
            </div>
          )}

          <div>
            <Text strong>Export Directory:</Text>
            <Space.Compact style={{ width: '100%', marginTop: 8 }}>
              <Input
                placeholder="Enter export directory path (e.g. /home/user/exports)"
                value={exportPath}
                onChange={(e) => setExportPath(e.target.value)}
                addonBefore={<FolderOutlined />}
                style={{ flex: 1 }}
              />
              <input
                type="file"
                {...({ webkitdirectory: "", directory: "" } as any)}
                multiple
                style={{ display: 'none' }}
                id="folder-picker"
                onChange={(e) => {
                  const files = e.target.files;
                  if (files && files.length > 0) {
                    // Get the directory path from the first file
                    const path = (files[0] as any).webkitRelativePath;
                    const folderPath = path.substring(0, path.lastIndexOf('/'));
                    if (folderPath) {
                      setExportPath(`/${folderPath}`);
                    }
                  }
                }}
              />
              <Button
                icon={<FolderOutlined />}
                onClick={() => {
                  const input = document.getElementById('folder-picker') as HTMLInputElement;
                  input?.click();
                }}
              >
                Browse
              </Button>
            </Space.Compact>
          </div>

          <div>
            <Text strong>Export Name:</Text>
            <Input
              placeholder={exportType === 'database' ? 'Enter export name prefix' : 'Enter file name'}
              value={exportName}
              onChange={(e) => setExportName(e.target.value)}
              style={{ marginTop: 8 }}
              addonAfter={exportType === 'database' ? `_[collection_name].${exportFormat}` : `.${exportFormat}`}
            />
            <Text type="secondary" style={{ fontSize: '12px', display: 'block', marginTop: 4 }}>
              {exportType === 'database' ? (
                <>Files will be saved as: <Text code>{exportName || 'export'}_[collection_name].{exportFormat}</Text></>
              ) : (
                <>File will be saved as: <Text code>{exportName || 'export'}.{exportFormat}</Text></>
              )}
            </Text>
          </div>

          <div>
            <Text strong>Export Format:</Text>
            <Select
              value={exportFormat}
              onChange={setExportFormat}
              style={{ width: '100%', marginTop: 8 }}
            >
              <Select.Option value="json">JSON</Select.Option>
              <Select.Option value="csv">CSV</Select.Option>
            </Select>
          </div>

          <div style={{ background: '#f6f6f6', padding: '12px', borderRadius: '6px' }}>
            <Text strong style={{ color: '#1890ff' }}>
              📋 Export Summary:
            </Text>
            <div style={{ marginTop: 8 }}>
              <Text>• Database: <Text code>{selectedDatabase}</Text></Text><br/>
              {exportType === 'database' ? (
                <>
                  <Text>• Collections: <Text strong>{collections?.length || 0}</Text></Text><br/>
                  <Text>• Total Documents: <Text strong>{stats?.total_documents || 0}</Text></Text><br/>
                </>
              ) : (
                <>
                  <Text>• Collection: <Text code>{selectedExportCollection}</Text></Text><br/>
                  <Text>• Documents: <Text strong>{collections.find(c => c.name === selectedExportCollection)?.document_count || 0}</Text></Text><br/>
                </>
              )}
              <Text>• Format: <Text code>{exportFormat.toUpperCase()}</Text></Text>
            </div>
          </div>

          <div style={{ background: '#fff7e6', padding: '12px', borderRadius: '6px', border: '1px solid #ffd591' }}>
            <Text style={{ color: '#d46b08' }}>
              ⚠️ <strong>Note:</strong> {exportType === 'database' ? 
                'Each collection will be exported as a separate file.' : 
                'The selected collection will be exported as a single file.'
              } Make sure the target directory exists and you have write permissions.
            </Text>
          </div>
        </Space>
      </Modal>
    </Space>
  );
};

export default DatabaseManager;
