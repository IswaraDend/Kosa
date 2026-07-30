import type { ReactNode } from 'react';

interface TableCardProps {
  title: string;
  count?: number;
  action?: ReactNode;
  children: ReactNode;
}

const TableCard = ({ title, count, action, children }: TableCardProps) => {
  return (
    <div className="table-card">
      <div className="table-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3>
          {title}
          {typeof count === 'number' ? ` (${count})` : ''}
        </h3>
        {action}
      </div>
      <table>{children}</table>
    </div>
  );
};

export default TableCard;
