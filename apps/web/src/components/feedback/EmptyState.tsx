import React from 'react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description: string;
  action?: React.ReactNode;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
}) => {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        textAlign: 'center',
        padding: '3rem 1.5rem',
        borderRadius: '8px',
        border: '1px dashed var(--border-subtle)',
        backgroundColor: 'var(--brand-white)',
        color: 'var(--text-muted)',
      }}
    >
      <div style={{ color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
        {icon || <Inbox size={32} strokeWidth={1.5} />}
      </div>
      <h3
        style={{
          fontSize: '1rem',
          fontWeight: 600,
          color: 'var(--text-primary)',
          margin: '0 0 0.25rem 0',
        }}
      >
        {title}
      </h3>
      <p
        style={{
          fontSize: '0.875rem',
          color: 'var(--text-secondary)',
          maxWidth: '400px',
          margin: '0 0 1rem 0',
          lineHeight: 1.4,
        }}
      >
        {description}
      </p>
      {action && <div>{action}</div>}
    </div>
  );
};
