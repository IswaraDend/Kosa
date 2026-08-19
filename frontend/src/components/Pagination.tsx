import { ChevronLeft, ChevronRight } from 'lucide-react';
import type { PageMeta } from '../hooks/usePagination';

interface PaginationProps {
  page: number;
  meta: PageMeta;
  onChange: (page: number) => void;
  /** Noun for the row count, e.g. "gudang" renders as "12 gudang". */
  label?: string;
}

/**
 * Renders nothing when everything fits on one page — a lone "Halaman 1 dari 1"
 * with two dead buttons is noise, and every list here shows its own count in
 * the table header already.
 */
const Pagination = ({ page, meta, onChange, label = 'data' }: PaginationProps) => {
  if (meta.totalPages <= 1) return null;

  return (
    <div className="pagination">
      <span className="pagination-info">
        Halaman {page} dari {meta.totalPages} — {meta.total} {label}
      </span>
      <div className="pagination-controls">
        <button className="pagination-btn" disabled={page <= 1} onClick={() => onChange(page - 1)}>
          <ChevronLeft size={15} />
          Sebelumnya
        </button>
        <button
          className="pagination-btn"
          disabled={page >= meta.totalPages}
          onClick={() => onChange(page + 1)}
        >
          Berikutnya
          <ChevronRight size={15} />
        </button>
      </div>
    </div>
  );
};

export default Pagination;
