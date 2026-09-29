import "../ui.css";
import "./auth.css";

/** Minimal shell for sign-in: the cream canvas, no nav, no footer. */
export default function AuthLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return <div className="ui auth-shell">{children}</div>;
}
