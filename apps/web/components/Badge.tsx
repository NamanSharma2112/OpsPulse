/** A coloured pill. The variant maps to a CSS class in globals.css. */
export function Badge({ variant, label }: { variant: string; label?: string }) {
  return (
    <span className={`badge badge-${variant}`}>{label ?? variant}</span>
  );
}
