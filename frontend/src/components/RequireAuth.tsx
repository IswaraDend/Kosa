import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

interface RequireAuthProps {
  children: ReactNode;
  requireRole?: 'super-admin' | 'admin' | 'member';
}

const RequireAuth = ({ children, requireRole }: RequireAuthProps) => {
  const { isAuthenticated, isSuperAdmin, isAdmin, isMember } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  const homePath = isSuperAdmin ? '/super-admin' : isAdmin ? '/admin' : isMember ? '/member' : '/login';

  if (requireRole === 'super-admin' && !isSuperAdmin) {
    return <Navigate to={homePath} replace />;
  }
  if (requireRole === 'admin' && !isSuperAdmin && !isAdmin) {
    return <Navigate to={homePath} replace />;
  }
  if (requireRole === 'member' && !isSuperAdmin && !isAdmin && !isMember) {
    return <Navigate to={homePath} replace />;
  }

  return <>{children}</>;
};

export default RequireAuth;
