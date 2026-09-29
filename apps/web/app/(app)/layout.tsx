import { redirect } from "next/navigation";
import { currentUser, isSignedIn } from "@/lib/api";
import { DashboardNav } from "@/components/DashboardNav";
import "../ui.css";
import "./dashboard.css";

/**
 * Shell for the signed-in product. Everything below this line is private, so
 * the gate lives here rather than being repeated in each page.
 */
export default async function DashboardLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  if (!(await isSignedIn())) redirect("/signin");
  const me = await currentUser();

  return (
    <div className="ui dash">
      <DashboardNav
        user={
          me.ok
            ? {
                name: me.data.name,
                email: me.data.email,
                login: me.data.github_login,
                avatar: me.data.avatar_url,
              }
            : null
        }
      />
      <main className="dash-main">
        <div className="ui-container">{children}</div>
      </main>
    </div>
  );
}
