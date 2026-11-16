import React from 'react';

interface Option {
  value: string;
  label: string;
}

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  options: Option[];
}

export const Select: React.FC<SelectProps> = ({
  label,
  options,
  style,
  id,
  ...props
}) => {
  const selectId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  return (
    <div style={{ display: 'inline-flex', flexDirection: 'column', gap: '0.25rem' }}>
      {label && (
        <label
          htmlFor={selectId}
          style={{
            fontSize: '0.75rem',
            fontWeight: 500,
            color: 'var(--text-muted)',
          }}
        >
          {label}
        </label>
      )}
      <select
        id={selectId}
        style={{
          padding: '0.45rem 0.75rem',
          fontSize: '0.8125rem',
          borderRadius: '6px',
          backgroundColor: 'var(--bg-card)',
          border: '1px solid var(--border-subtle)',
          color: 'var(--text-primary)',
          cursor: 'pointer',
          outline: 'none',
          ...style,
        }}
        {...props}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
};
