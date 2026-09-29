/** The OpsPulse mark. One definition, used by every surface. */
export function Wordmark({ size = 18 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 18 18" aria-hidden="true">
      <rect width="18" height="18" rx="5" fill="#111111" />
      <path
        d="M3.6 9.4h2.6l1.5-3.6 2 7 1.6-3.4h3.1"
        fill="none"
        stroke="#ffffff"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
