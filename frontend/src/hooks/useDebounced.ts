import { useEffect, useState } from 'react';

/**
 * Delays a rapidly-changing value — a search box, typically.
 *
 * Once searching moved server-side it happens per keystroke, so without this
 * every letter typed would fire a request and the replies could arrive out of
 * order, briefly showing results for a prefix the user already moved past.
 */
export function useDebounced<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}
