import "./design.css";
import "./style.css";
import "./board-premium.css";
import "./release.css";
import { mountApp } from "./ui/App";
import { bootGameApp } from "./ui/boot";

const app = document.querySelector<HTMLElement>("#app");
if (!app) throw new Error("Elemen #app tidak ada");

const path = location.pathname.replace(/\/+$/, "");
const admin = path === "/admin" || path.endsWith("/admin") || path.includes("/admin/") || new URLSearchParams(location.search).get("panel") === "admin";

if (admin) {
  document.body.classList.add("admin-panel");
  document.title = "Admin — Ular Tangga Nusantara";
  void import("./admin/admin").then((m) => m.mountAdmin(app));
} else {
  void bootGameApp(mountApp);
}
