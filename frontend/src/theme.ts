import { createTheme, responsiveFontSizes } from "@mui/material/styles";

let theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#3f51b5",
    },
    secondary: {
      main: "#606FC7",
    },
    background: {
      default: "#ffffff",
    },
  },
});

theme = responsiveFontSizes(theme);

export { theme };