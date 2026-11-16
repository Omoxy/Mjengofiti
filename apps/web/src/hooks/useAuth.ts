import { useState, useEffect, useCallback } from 'react';
import { api, type UserProfile } from '../api/client';

export interface RoleDefinition {
  role: string;
  title: string;
  description: string;
  permissions: string[];
}

export function useAuth() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [roles, setRoles] = useState<RoleDefinition[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAuth = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [profileRes, rolesRes] = await Promise.all([
        api.getUserProfile(),
        api.getRoles(),
      ]);
      setUser(profileRes);
      setRoles(rolesRes.roles || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load user profile');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAuth();
  }, [fetchAuth]);

  const hasPermission = useCallback(
    (permission: string): boolean => {
      if (!user) return false;
      if (user.permissions.includes('*')) return true;
      return user.permissions.includes(permission);
    },
    [user]
  );

  return {
    user,
    roles,
    isLoading,
    error,
    hasPermission,
    refetch: fetchAuth,
  };
}
