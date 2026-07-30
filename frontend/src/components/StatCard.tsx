import type { ReactNode } from 'react';

interface StatCardProps {
  label: string;
  value: string | number;
  icon: ReactNode;
}

const StatCard = ({ label, value, icon }: StatCardProps) => {
  return (
    <div className="stat-card">
      <div className="stat-header">
        <span>{label}</span>
        <span className="stat-icon">{icon}</span>
      </div>
      <div className="stat-value">{value}</div>
    </div>
  );
};

export default StatCard;
