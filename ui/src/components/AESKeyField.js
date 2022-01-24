import React, { Component } from "react";

import TextField from "@material-ui/core/TextField";
import InputAdornment from '@material-ui/core/InputAdornment';
import IconButton from '@material-ui/core/IconButton';
import Button from "@material-ui/core/Button";
import Tooltip from '@material-ui/core/Tooltip';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import Fade from '@material-ui/core/Fade';

import Eye from "mdi-material-ui/Eye";
import EyeOff from "mdi-material-ui/EyeOff";
import Refresh from "mdi-material-ui/Refresh";
import Copy from "mdi-material-ui/ContentCopy";

import MaskedInput from "react-text-mask";
import {successNotify, errorHandler } from "../stores/helpers";

function CopyMenu(props) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const open = Boolean(anchorEl);

  function handleClick(event) {
    setAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setAnchorEl(null);
  }

  function copyToClipboard() {
    const bytes = props.value.match(/[A-Fa-f0-9]{2}/g);

    if (bytes !== null && navigator.clipboard !== undefined) {
        navigator.clipboard.writeText(bytes.join("").toUpperCase()).then(successNotify("Copied to clipboard")).catch(errorHandler);
    }
    handleClose()
  }
  function copyToClipboardHexArray() {
    const bytes = props.value.match(/[A-Fa-f0-9]{2}/g);

    if (bytes !== null && navigator.clipboard !== undefined) {
      navigator.clipboard.writeText(bytes.join(", ").toUpperCase().replace(/[A-Fa-f0-9]{2}/g, "0x$&")).then(successNotify("Copied to clipboard")).catch(errorHandler);
    }
    handleClose()
  }

  return (
    <div>
      <Tooltip title="Click to copy">
        <IconButton
              aria-controls="fade-menu"
              aria-haspopup="true"
              onClick={handleClick}
              aria-label="Toggle key visibility"
              >
              <Copy />
          </IconButton>
        </Tooltip>
      <Menu
        id="fade-menu"
        anchorEl={anchorEl}
        keepMounted
        open={open}
        onClose={handleClose}
        TransitionComponent={Fade}
      >
        <MenuItem onClick={copyToClipboard}>Hex String</MenuItem>
        <MenuItem onClick={copyToClipboardHexArray}>Hex Array</MenuItem>
      </Menu>
    </div>
  );
}

class AESKeyHEXMask extends Component {
  render() {
    const { inputRef, ...other } = this.props;

    return(
      <MaskedInput
        {...other}
        ref={(ref) => {
          inputRef(ref ? ref.inputElement : null);
        }}
        mask={[
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
          ' ',
          /[A-Fa-f0-9]/,
          /[A-Fa-f0-9]/,
        ]}
      />
    );
  }
}


class AESKeyField extends Component {
  constructor() {
    super();

    this.state = {
      showKey: false,
      msb: true,
      value: "",
    };
  }

  toggleShowPassword = () => {
    this.setState({
      showKey: !this.state.showKey,
    });
  }

  toggleByteOrder = () => {
    this.setState({
      msb: !this.state.msb,
    });

    const bytes = this.state.value.match(/[A-Fa-f0-9]{2}/g);
    if (bytes !== null) {
      this.setState({
        value: bytes.reverse().join(" "),
      });
    }
  }

  randomKey = () => {
    let cryptoObj = window.crypto || window.msCrypto;
    let b = new Uint8Array(16);
    cryptoObj.getRandomValues(b);

    let key = Buffer.from(b).toString('hex');
    this.setState({
      value: key,
    });

    let str = "";
    const bytes = key.match(/[A-Fa-f0-9]{2}/g);
    if (!this.state.msb && bytes !== null) {
      str = bytes.reverse().join("");
    } else if (bytes !== null) {
      str = bytes.join("");
    } 

    this.props.onChange({
      target: {
        value: str,
        type: "text",
        id: this.props.id,
      },
    });
  }

  onChange = (e) => {
    this.setState({
      value: e.target.value,
    });

    let str = "";

    const bytes = e.target.value.match(/[A-Fa-f0-9]{2}/g);
    if (!this.state.msb && bytes !== null) {
      str = bytes.reverse().join("");
    } else if (bytes !== null) {
      str = bytes.join("");
    } 

    this.props.onChange({
      target: {
        value: str,
        type: "text",
        id: this.props.id,
      },
    });
  }

  componentDidMount() {
    this.setState({
      value: this.props.value || "",
      showKey: this.props.value === "" ? true : false,
    });
  }

  render() {
    return(
      <TextField
        type={this.state.showKey ? "text" : "password"}
        InputProps={{
          inputComponent: AESKeyHEXMask,
          endAdornment: <InputAdornment position="end">
            {this.state.showKey && <Tooltip title="Toggle the byte order of the input. Some devices use LSB.">
              <Button
                aria-label="Toggle byte order"
                onClick={this.toggleByteOrder}
              >
                {this.state.msb ? "MSB": "LSB"}
              </Button>
            </Tooltip>}
            {this.props.random && this.state.showKey && !this.props.disabled && <Tooltip title="Generate random key.">
              <IconButton
                aria-label="Generate random key"
                onClick={this.randomKey}
              >
                <Refresh />
              </IconButton>
             </Tooltip>}
            {this.state.showKey &&
              <CopyMenu
                value={this.state.value}
              />}
            <IconButton
              aria-label="Toggle key visibility"
              onClick={this.toggleShowPassword}
            >
              {this.state.showKey ? <EyeOff /> : <Eye />}
            </IconButton>
          </InputAdornment>
        }}
        {...this.props}
        onChange={this.onChange}
        value={this.state.value}
        disabled={this.props.disabled || !this.state.showKey}
      />
    );
  }
}

export default AESKeyField;
