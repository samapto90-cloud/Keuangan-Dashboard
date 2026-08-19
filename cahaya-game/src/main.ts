import "./style.css";
import { mountApp } from "./ui/App";

const app = document.querySelector<HTMLElement>("#app");
if (!app) throw new Error("Elemen #app tidak ada");
mountApp(app);
