import { createMuiTheme } from "@material-ui/core/styles";
import blue from "@material-ui/core/colors/blue";


const theme = createMuiTheme({
    palette: {
      primary: {
        light: blue[400],
        main: blue[500],
        dark: blue[600],
      },
    },
});
  
export default theme;
