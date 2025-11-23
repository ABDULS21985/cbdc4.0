import React from 'react';

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
    label?: string;
    error?: string;
    options: { value: string; label: string }[];
    placeholder?: string;
}

export const Select: React.FC<SelectProps> = ({
    label,
    error,
    options,
    placeholder,
    className,
    ...props
}) => {
    return (
        <div className="flex flex-col gap-1">
            {label && (
                <label className="text-sm font-medium text-gray-700">{label}</label>
            )}
            <select
                className={`border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-green-500 bg-white ${
                    error ? 'border-red-500' : 'border-gray-300'
                } ${className || ''}`}
                {...props}
            >
                {placeholder && (
                    <option value="" disabled>
                        {placeholder}
                    </option>
                )}
                {options.map((option) => (
                    <option key={option.value} value={option.value}>
                        {option.label}
                    </option>
                ))}
            </select>
            {error && <span className="text-xs text-red-500">{error}</span>}
        </div>
    );
};
