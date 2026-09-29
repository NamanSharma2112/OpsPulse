/**
 * A state pill. The dot carries the colour, the label stays charcoal — so
 * colour encodes the state without costing the text its contrast.
 */
export function Badge({ variant, label }: { variant: string; label?: string }) {
  return <span className={`ui-badge ui-badge-${variant}`}>{label ?? variant}</span>;
}
