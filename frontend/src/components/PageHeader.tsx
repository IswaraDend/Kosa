import type { ReactNode } from 'react';
import { Search } from 'lucide-react';

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  actions?: ReactNode;
}

const PageHeader = ({
  title,
  subtitle,
  searchValue,
  onSearchChange,
  searchPlaceholder = 'Cari...',
  actions,
}: PageHeaderProps) => {
  return (
    <div className="page-header">
      <div className="page-title">
        <h1>{title}</h1>
        {subtitle && <p>{subtitle}</p>}
      </div>

      <div className="header-actions">
        {onSearchChange && (
          <div className="search-box">
            <Search size={16} />
            <input
              type="text"
              placeholder={searchPlaceholder}
              value={searchValue}
              onChange={(e) => onSearchChange(e.target.value)}
            />
          </div>
        )}
        {actions}
      </div>
    </div>
  );
};

export default PageHeader;
