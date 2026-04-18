import { useEffect, useRef, useState } from "react";

interface AdaptiveGridPageSizeOptions {
  enabled: boolean;
  defaultPageSize: number;
  targetRows?: number;
  minRows?: number;
  maxRows?: number;
}

export function useAdaptiveGridPageSize({
  enabled,
  defaultPageSize,
  targetRows = 4,
  minRows = 1,
  maxRows = 4,
}: AdaptiveGridPageSizeOptions) {
  const gridRef = useRef<HTMLDivElement | null>(null);
  const paginationRef = useRef<HTMLDivElement | null>(null);
  const [pageSize, setPageSize] = useState(defaultPageSize);

  useEffect(() => {
    if (!enabled) {
      setPageSize(defaultPageSize);
      return;
    }

    let frameId = 0;
    let resizeObserver: ResizeObserver | null = null;

    const measure = () => {
      cancelAnimationFrame(frameId);
      frameId = window.requestAnimationFrame(() => {
        const gridElement = gridRef.current;
        if (!gridElement) return;

        const computedStyle = window.getComputedStyle(gridElement);
        const columns = computedStyle.gridTemplateColumns
          .split(" ")
          .filter(Boolean).length;

        if (!columns) return;

        const rows = Math.max(minRows, Math.min(targetRows, maxRows));
        const nextPageSize = Math.max(columns * rows, columns);

        setPageSize((current) => (current === nextPageSize ? current : nextPageSize));
      });
    };

    measure();
    window.addEventListener("resize", measure);

    if (typeof ResizeObserver !== "undefined") {
      resizeObserver = new ResizeObserver(() => measure());
      if (gridRef.current) resizeObserver.observe(gridRef.current);
      if (paginationRef.current) resizeObserver.observe(paginationRef.current);
      const mainElement = gridRef.current?.closest("main");
      if (mainElement) resizeObserver.observe(mainElement);
    }

    return () => {
      cancelAnimationFrame(frameId);
      window.removeEventListener("resize", measure);
      resizeObserver?.disconnect();
    };
  }, [defaultPageSize, enabled, maxRows, minRows, targetRows]);

  return { gridRef, pageSize, paginationRef };
}