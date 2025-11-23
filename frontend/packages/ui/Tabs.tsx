import React, { createContext, useContext, useState } from 'react';

interface TabsContextType {
    activeTab: string;
    setActiveTab: (id: string) => void;
}

const TabsContext = createContext<TabsContextType | null>(null);

interface TabsProps {
    children: React.ReactNode;
    defaultTab?: string;
    onChange?: (tabId: string) => void;
}

export const Tabs: React.FC<TabsProps> = ({ children, defaultTab, onChange }) => {
    const [activeTab, setActiveTab] = useState(defaultTab || '');

    const handleTabChange = (id: string) => {
        setActiveTab(id);
        onChange?.(id);
    };

    return (
        <TabsContext.Provider value={{ activeTab, setActiveTab: handleTabChange }}>
            <div>{children}</div>
        </TabsContext.Provider>
    );
};

export const TabList: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
    <div className={`flex border-b ${className || ''}`}>
        {children}
    </div>
);

interface TabProps {
    id: string;
    children: React.ReactNode;
}

export const Tab: React.FC<TabProps> = ({ id, children }) => {
    const context = useContext(TabsContext);
    if (!context) throw new Error('Tab must be used within Tabs');

    const { activeTab, setActiveTab } = context;
    const isActive = activeTab === id;

    return (
        <button
            onClick={() => setActiveTab(id)}
            className={`px-4 py-3 font-medium text-sm border-b-2 transition-colors ${
                isActive
                    ? 'border-green-600 text-green-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
        >
            {children}
        </button>
    );
};

interface TabPanelProps {
    id: string;
    children: React.ReactNode;
}

export const TabPanel: React.FC<TabPanelProps> = ({ id, children }) => {
    const context = useContext(TabsContext);
    if (!context) throw new Error('TabPanel must be used within Tabs');

    if (context.activeTab !== id) return null;

    return <div className="py-4">{children}</div>;
};
