import React from 'react';
import { AlertTriangle, AlertCircle, RefreshCw } from 'lucide-react';

interface ErrorAlertProps {
  title?: string;
  message: string;
  variant?: 'error' | 'warning' | 'degraded';
  onRetry?: () => void;
  className?: string;
}

export const ErrorAlert: React.FC<ErrorAlertProps> = ({
  title,
  message,
  variant = 'error',
  onRetry,
  className = '',
}) => {
  const isDegraded = variant === 'degraded';
  const isWarning = variant === 'warning';

  const borderColor = variant === 'error' ? 'var(--gis-crimson-border)' : 'var(--gis-amber-border)';
  const bgColor = variant === 'error' ? 'var(--gis-crimson-subtle)' : 'var(--gis-amber-subtle)';
  const textColor = variant === 'error' ? 'var(--brand-maroon-deep)' : 'var(--brand-maroon)';
  const Icon = isDegraded || isWarning ? AlertTriangle : AlertCircle;

  return (
    <div
      role="alert"
      className={`error-alert ${className}`}
      style={{
        display: 'flex',
        alignItems: 'flex-start',
        gap: '0.75rem',
        padding: '0.875rem 1rem',
        borderRadius: '6px',
        backgroundColor: bgColor,
        border: `1px solid ${borderColor}`,
        color: 'var(--text-primary)',
        margin: '0.5rem 0',
      }}
    >
      <Icon size={18} color={borderColor} style={{ flexShrink: 0, marginTop: '2px' }} />
      <div style={{ flex: 1 }}>
        {title && (
          <div style={{ fontWeight: 600, fontSize: '0.875rem', color: textColor, marginBottom: '0.25rem' }}>
            {title}
          </div>
        )}
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)', lineHeight: 1.4 }}>
          {message}
        </div>
      </div>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.25rem',
            padding: '0.375rem 0.625rem',
            borderRadius: '4px',
            backgroundColor: 'transparent',
            border: `1px solid ${borderColor}`,
            color: textColor,
            fontSize: '0.75rem',
            cursor: 'pointer',
            flexShrink: 0,
          }}
        >
          <RefreshCw size={12} />
          Retry
        </button>
      )}
    </div>
  );
};
