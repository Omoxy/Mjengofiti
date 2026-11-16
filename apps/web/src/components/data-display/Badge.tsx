import React from 'react';

export type BadgeVariant =
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'neutral'
  | 'primary';

interface BadgeProps {
  children: React.ReactNode;
  variant?: BadgeVariant;
  size?: 'sm' | 'md';
  className?: string;
  style?: React.CSSProperties;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'neutral',
  size = 'md',
  className = '',
  style,
}) => {
  const getColors = () => {
    switch (variant) {
      case 'success':
        return { bg: 'var(--brand-cream)', text: 'var(--brand-maroon)', border: 'var(--brand-border)' };
      case 'warning':
        return { bg: 'var(--brand-cream)', text: 'var(--brand-maroon)', border: 'var(--brand-border)' };
      case 'danger':
        return { bg: 'var(--brand-maroon-soft)', text: 'var(--brand-maroon-deep)', border: 'var(--gis-crimson-border)' };
      case 'info':
        return { bg: 'var(--brand-cream-light)', text: 'var(--brand-maroon)', border: 'var(--brand-border)' };
      case 'primary':
        return { bg: 'var(--brand-maroon-soft)', text: 'var(--brand-maroon)', border: 'var(--gis-cyan-border)' };
      case 'neutral':
      default:
        return { bg: 'var(--brand-white)', text: 'var(--text-secondary)', border: 'var(--border-subtle)' };
    }
  };

  const colors = getColors();
  const padding = size === 'sm' ? '0.125rem 0.375rem' : '0.2rem 0.5rem';
  const fontSize = size === 'sm' ? '0.6875rem' : '0.75rem';

  return (
    <span
      className={`badge ${className}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        padding,
        fontSize,
        fontWeight: 600,
        lineHeight: 1,
        borderRadius: '9999px',
        backgroundColor: colors.bg,
        color: colors.text,
        border: `1px solid ${colors.border}`,
        textTransform: 'capitalize',
        ...style,
      }}
    >
      {children}
    </span>
  );
};
