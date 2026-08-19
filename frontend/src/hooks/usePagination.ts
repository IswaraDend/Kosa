import { useEffect, useState } from 'react';

/** Rows per page requested from the API. Mirrors defaultPerPage in the Go side. */
export const PER_PAGE = 25;

export interface PageMeta {
  total: number;
  totalPages: number;
}

/**
 * Page state for a server-paginated list.
 *
 * `resetKey` is whatever narrows the result set — the selected project, a type
 * filter, a search term. When it changes, the current page number stops
 * meaning anything (page 4 of the previous result set may not exist in the new
 * one, leaving the user staring at an empty table), so it snaps back to 1.
 */
export function usePagination(resetKey: unknown) {
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState<PageMeta>({ total: 0, totalPages: 1 });

  useEffect(() => {
    setPage(1);
  }, [resetKey]);

  return { page, setPage, meta, setMeta };
}

/** Reads the envelope fields that listResponse() returns on the Go side. */
export function metaFrom(res: { total?: number; total_pages?: number }): PageMeta {
  return { total: res.total ?? 0, totalPages: res.total_pages ?? 1 };
}
