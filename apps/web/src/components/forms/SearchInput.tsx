import React, { useEffect, useState } from 'react';
import { Search, X } from 'lucide-react';

interface SearchInputProps {
  value?: string;
  placeholder?: string;
  onChange?: (val: string) => void;
  debounceMs?: number;
  className?: string;
}

export const SearchInput: React.FC<SearchInputProps> = ({
  value: initialValue = '',
  placeholder = 'Search...',
  onChange,
  debounceMs = 250,
  className = '',
}) => {
  const [query, setQuery] = useState(initialValue);
  const [prevInitialValue, setPrevInitialValue] = useState(initialValue);

  if (initialValue !== prevInitialValue) {
    setPrevInitialValue(initialValue);
    setQuery(initialValue);
  }


  useEffect(() => {
    const handler = setTimeout(() => {
      onChange?.(query);
    }, debounceMs);
    return () => clearTimeout(handler);
  }, [query, debounceMs, onChange]);

  return (
    <div
      className={`search-input-wrapper ${className}`}
      style={{
        position: 'relative',
        display: 'flex',
        alignItems: 'center',
        width: '100%',
      }}
    >
      <Search
        size={14}
        color="var(--text-muted)"
        style={{
          position: 'absolute',
          left: '0.625rem',
          pointerEvents: 'none',
        }}
      />
      <input
        type="text"
        value={query}
        placeholder={placeholder}
        onChange={(e) => setQuery(e.target.value)}
        style={{
          width: '100%',
          padding: '0.45rem 1.75rem 0.45rem 2rem',
          fontSize: '0.8125rem',
          borderRadius: '6px',
          backgroundColor: 'var(--bg-card)',
          border: '1px solid var(--border-subtle)',
          color: 'var(--text-primary)',
          outline: 'none',
        }}
        aria-label={placeholder}
      />
      {query && (
        <button
          type="button"
          onClick={() => {
            setQuery('');
            onChange?.('');
          }}
          aria-label="Clear search"
          style={{
            position: 'absolute',
            right: '0.5rem',
            background: 'none',
            border: 'none',
            color: 'var(--text-muted)',
            cursor: 'pointer',
            padding: 0,
            display: 'flex',
            alignItems: 'center',
          }}
        >
          <X size={14} />
        </button>
      )}
    </div>
  );
};
