import React from 'react';

export const Card: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => {
    return (
        <div className={`bg-white shadow-md rounded-lg p-6 ${className || ''}`}>
            {children}
        </div>
    );
};
