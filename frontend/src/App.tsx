import React, { useState, useEffect, useCallback } from 'react';
import { Layout, Typography, Space, ConfigProvider, Button, Dropdown, Menu, App as AntdApp, Tabs } from 'antd';
import { DatabaseOutlined, UserOutlined, LogoutOutlined, FileTextOutlined, ApiOutlined } from '@ant-design/icons';
import DatabaseManager from './components/DatabaseManager';
import DocumentManager from './components/DocumentManager';
import ApiManager from './components/ApiManager';
import LoginForm from './components/LoginForm';
import { ValidateSession } from '../wailsjs/go/main/App';
import './App.css';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

interface User {
    id: string;
    username: string;
    email: string;
}

function App() {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [sessionId, setSessionId] = useState<string | null>(null);
    const [currentUser, setCurrentUser] = useState<User | null>(null);
    const [selectedDatabase, setSelectedDatabase] = useState<string | null>(null);
    const [selectedCollection, setSelectedCollection] = useState<string | null>(null);
    const [collections, setCollections] = useState<any[]>([]);
    const [refreshDatabaseManager, setRefreshDatabaseManager] = useState<() => Promise<void>>();
    const [updateDatabaseStats, setUpdateDatabaseStats] = useState<() => void>();
    const [removeCollectionFromDbManager, setRemoveCollectionFromDbManager] = useState<(collectionName: string) => void>();

    // Check for existing session on app start
    useEffect(() => {
        const storedSessionId = localStorage.getItem('enginenosql_session');
        if (storedSessionId) {
            validateStoredSession(storedSessionId);
        }
    }, []);

    const validateStoredSession = async (storedSessionId: string) => {
        try {
            const session = await ValidateSession(storedSessionId);
            if (session && session.is_active) {
                setSessionId(storedSessionId);
                setCurrentUser({
                    id: session.user_id,
                    username: session.username,
                    email: '' // We don't have email in session, will be loaded separately
                });
                setIsAuthenticated(true);
            } else {
                // Session invalid, remove from storage
                localStorage.removeItem('enginenosql_session');
            }
        } catch (err) {
            console.error('Session validation failed:', err);
            localStorage.removeItem('enginenosql_session');
        }
    };

    useEffect(() => {
        if (selectedDatabase && sessionId) {
            loadCollections();
        } else {
            setCollections([]);
            setSelectedCollection(null);
        }
    }, [selectedDatabase, sessionId]);

    const loadCollections = async () => {
        if (!selectedDatabase || !sessionId) return;
        
        try {
            // Collections will be loaded by DatabaseManager component
            // and passed to DocumentManager via props
        } catch (err) {
            console.error('Failed to load collections:', err);
        }
    };

    const handleLoginSuccess = (newSessionId: string, user: User) => {
        setSessionId(newSessionId);
        setCurrentUser(user);
        setIsAuthenticated(true);
        localStorage.setItem('enginenosql_session', newSessionId);
    };

    const handleLogout = async () => {
        try {
            if (sessionId) {
                // TODO: Call Logout with sessionId when bindings are updated
                // await Logout(sessionId);
            }
        } catch (err) {
            console.error('Logout error:', err);
        } finally {
            setSessionId(null);
            setCurrentUser(null);
            setIsAuthenticated(false);
            setSelectedDatabase(null);
            setSelectedCollection(null);
            setCollections([]);
            localStorage.removeItem('enginenosql_session');
        }
    };

    const handleDatabaseSelect = (database: string) => {
        setSelectedDatabase(database);
        setSelectedCollection(null);
    };

    const handleCollectionSelect = (collection: string) => {
        setSelectedCollection(collection || '');
    };

    // Prosta funkcja aktualizująca kolekcje - rezygnujemy z nadmiernej optymalizacji
    const handleCollectionsChange = useCallback((newCollections: any[]) => {
        setCollections(newCollections);
    }, []);

    // Funkcja wywoływana, gdy kolekcja została usunięta - synchronizuje wszystkie komponenty
    const handleCollectionDeleted = async (deletedCollectionName: string) => {
        try {
            // Natychmiast usuń kolekcję ze stanu głównego App
            setCollections(prevCollections => 
                prevCollections.filter(col => col.name !== deletedCollectionName)
            );
            
            // Jeśli usunięta kolekcja była wybrana, wyczyść wybór
            if (selectedCollection === deletedCollectionName) {
                setSelectedCollection('');
            }

            // Aktualizacja statystyk bazy danych
            if (updateDatabaseStats && typeof updateDatabaseStats === 'function') {
                updateDatabaseStats();
            }

            // Usuń kolekcję z listy DatabaseManager
            if (removeCollectionFromDbManager && typeof removeCollectionFromDbManager === 'function') {
                removeCollectionFromDbManager(deletedCollectionName);
            }
            
            // Odśwież dane - bez warunku, bo to zawsze jest prawdziwe gdy jesteśmy w tym miejscu
            await refreshAllData();
            
        } catch (err) {
            console.error('Błąd podczas obsługi usunięcia kolekcji:', err);
            // W razie błędu, ustawiamy pustą tablicę 
            setCollections([]);
        }
    };

    const refreshAllData = async () => {
        // Refresh all application data
        if (refreshDatabaseManager && typeof refreshDatabaseManager === 'function') {
            try {
                await refreshDatabaseManager();
            } catch (err) {
                console.error('Error refreshing database manager:', err);
            }
        }
        
        // Force immediate update of collections state
        if (selectedDatabase && sessionId) {
            try {
                // Import the function directly to get fresh data
                const { GetCollections } = await import('../wailsjs/go/main/App');
                const freshCollections = await GetCollections(sessionId, selectedDatabase);
                const collections = freshCollections || [];
                setCollections(collections);
            } catch (err) {
                console.error('Error loading fresh collections:', err);
                setCollections([]);
            }
        }
    };

    if (!isAuthenticated) {
        return <LoginForm onLoginSuccess={handleLoginSuccess} />;
    }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: (
        <div>
          <div><strong>{currentUser?.username}</strong></div>
          <div style={{ fontSize: '12px', color: '#666' }}>{currentUser?.email}</div>
        </div>
      )
    },
    {
      type: 'divider' as const
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: handleLogout,
      danger: true
    }
  ];

    return (
        <ConfigProvider
            theme={{
                token: {
                    colorPrimary: '#1890ff',
                },
            }}
        >
            <AntdApp>
                <Layout style={{ minHeight: '100vh' }}>
                <Header style={{ 
                    background: '#001529', 
                    display: 'flex', 
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0 24px'
                }}>
                    <Space>
                        <DatabaseOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
                        <Title level={3} style={{ margin: 0, color: 'white' }}>
                            EngineNoSQL - Professional NoSQL Database
                        </Title>
                    </Space>
                    
                    <Space>
                        <Text style={{ color: 'white' }}>
                            Welcome, {currentUser?.username}
                        </Text>
                        <Dropdown menu={{ items: userMenuItems }} trigger={['click']}>
                            <Button 
                                type="text" 
                                style={{ color: 'white' }}
                                icon={<UserOutlined />}
                            />
                        </Dropdown>
                    </Space>
                </Header>
                
                <Content style={{ padding: '24px', background: '#f0f2f5' }}>
                    <Tabs
                        defaultActiveKey="database"
                        size="large"
                        items={[
                            {
                                key: 'database',
                                label: (
                                    <Space>
                                        <DatabaseOutlined />
                                        Database Manager
                                    </Space>
                                ),
                                children: (
                                    <DatabaseManager 
                                        onDatabaseSelect={handleDatabaseSelect}
                                        selectedDatabase={selectedDatabase}
                                        sessionId={sessionId}
                                        onCollectionsChange={handleCollectionsChange}
                                        onRefreshNeeded={setRefreshDatabaseManager}
                                        onStatsUpdateNeeded={setUpdateDatabaseStats}
                                        onCollectionRemovalNeeded={setRemoveCollectionFromDbManager}
                                    />
                                )
                            },
                            {
                                key: 'documents',
                                label: (
                                    <Space>
                                        <FileTextOutlined />
                                        Document Manager
                                    </Space>
                                ),
                                children: (
                                    <DocumentManager
                                        selectedDatabase={selectedDatabase}
                                        selectedCollection={selectedCollection}
                                        collections={collections}
                                        onCollectionSelect={handleCollectionSelect}
                                        sessionId={sessionId}
                                        onDataChange={refreshAllData}
                                        onCollectionDeleted={handleCollectionDeleted}
                                    />
                                )
                            },
                            {
                                key: 'api',
                                label: (
                                    <Space>
                                        <ApiOutlined />
                                        API Manager
                                    </Space>
                                ),
                                children: (
                                    <ApiManager
                                        selectedDatabase={selectedDatabase}
                                        sessionId={sessionId}
                                    />
                                )
                            }
                        ]}
                    />
                </Content>
                </Layout>
            </AntdApp>
        </ConfigProvider>
    );
}

export default App;
