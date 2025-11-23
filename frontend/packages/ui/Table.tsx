import React from 'react';

interface TableProps {
    children: React.ReactNode;
    className?: string;
}

export const Table: React.FC<TableProps> = ({ children, className }) => (
    <div className={`overflow-x-auto ${className || ''}`}>
        <table className="w-full">
            {children}
        </table>
    </div>
);

export const TableHead: React.FC<TableProps> = ({ children }) => (
    <thead className="bg-gray-50 border-b">
        {children}
    </thead>
);

export const TableBody: React.FC<TableProps> = ({ children }) => (
    <tbody className="divide-y">
        {children}
    </tbody>
);

export const TableRow: React.FC<TableProps & { hover?: boolean }> = ({ children, hover = true }) => (
    <tr className={hover ? 'hover:bg-gray-50' : ''}>
        {children}
    </tr>
);

export const TableHeader: React.FC<TableProps> = ({ children, className }) => (
    <th className={`px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${className || ''}`}>
        {children}
    </th>
);

export const TableCell: React.FC<TableProps> = ({ children, className }) => (
    <td className={`px-4 py-3 text-sm ${className || ''}`}>
        {children}
    </td>
);
