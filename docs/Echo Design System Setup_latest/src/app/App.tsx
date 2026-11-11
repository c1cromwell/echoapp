import { RouterProvider } from "react-router";
import { router } from "./routes";
import { ThemeProvider } from "./components/echo-theme";
import { HiddenFolderProvider } from "./components/hidden-folder-data";

export default function App() {
  return (
    <ThemeProvider>
      <HiddenFolderProvider>
        <RouterProvider router={router} />
      </HiddenFolderProvider>
    </ThemeProvider>
  );
}