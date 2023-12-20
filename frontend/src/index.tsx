import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import reportWebVitals from "./reportWebVitals";
import { persistor, store } from "./store/store";
import { Provider } from "react-redux";
import UserAlert from "./components/UserAlert";
import { ThemeProvider } from "@mui/material/styles";
import { theme } from "./theme";
import { PersistGate } from "redux-persist/integration/react";
import { CssBaseline } from "@mui/material";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <App />
          <UserAlert />
        </ThemeProvider>
      </PersistGate>
    </Provider>
  </React.StrictMode>
);

reportWebVitals();
