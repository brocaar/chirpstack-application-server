import React, { Component } from "react";

import { withStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import Input from '@material-ui/core/Input';
import MenuItem from '@material-ui/core/MenuItem';
import Chip from '@material-ui/core/Chip';
import FormControl from "@material-ui/core/FormControl";

import MenuDown from "mdi-material-ui/MenuDown";
import Cancel from "mdi-material-ui/Cancel";
import MenuUp from "mdi-material-ui/MenuUp";
import Close from "mdi-material-ui/Close";
import Select from 'react-select';
import 'react-select/dist/react-select.css';


// taken from react-select example
// https://material-ui.com/demos/autocomplete/


const ITEM_HEIGHT = 48;

const styles = theme => ({
  chip: {
    margin: theme.spacing.unit / 4,
  },
  '@global': {
    '.Select-control': {
      display: 'flex',
      alignItems: 'center',
      border: 0,
      height: 'auto',
      background: 'transparent',
      '&:hover': {
        boxShadow: 'none',
      },
    },
    '.Select-multi-value-wrapper': {
      flexGrow: 1,
      display: 'flex',
      flexWrap: 'wrap',
    },
    '.Select--multi .Select-input': {
      margin: 0,
    },
    '.Select.has-value.is-clearable.Select--single > .Select-control .Select-value': {
      padding: 0,
    },
    '.Select-noresults': {
      padding: theme.spacing.unit * 2,
    },
    '.Select-input': {
      display: 'inline-flex !important',
      padding: 0,
      height: 'auto',
    },
    '.Select-input input': {
      background: 'transparent',
      border: 0,
      padding: 0,
      cursor: 'default',
      display: 'inline-block',
      fontFamily: 'inherit',
      fontSize: 'inherit',
      margin: 0,
      outline: 0,
    },
    '.Select-placeholder, .Select--single .Select-value': {
      position: 'absolute',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      display: 'flex',
      alignItems: 'center',
      fontFamily: theme.typography.fontFamily,
      fontSize: theme.typography.pxToRem(16),
      padding: 0,
    },
    '.Select-value': {
      color: "black !important",
      paddingLeft: "0 !important",
    },
    '.Select-placeholder': {
      opacity: 0.42,
      color: theme.palette.common.black,
    },
    '.Select-menu-outer': {
      backgroundColor: theme.palette.background.paper,
      boxShadow: theme.shadows[2],
      position: 'absolute',
      left: 0,
      top: `calc(100% + ${theme.spacing.unit}px)`,
      width: '100%',
      zIndex: 2,
      maxHeight: ITEM_HEIGHT * 4.5,
    },
    '.Select.is-focused:not(.is-open) > .Select-control': {
      boxShadow: 'none',
    },
    '.Select-menu': {
      maxHeight: ITEM_HEIGHT * 4.5,
      overflowY: 'auto',
      zIndex: 99999,
    },
    '.Select-menu div': {
      boxSizing: 'content-box',
    },
    '.Select-arrow-zone, .Select-clear-zone': {
      color: theme.palette.action.active,
      cursor: 'pointer',
      height: 21,
      width: 21,
      zIndex: 1,
    },
    // Only for screen readers. We can't use display none.
    '.Select-aria-only': {
      position: 'absolute',
      overflow: 'hidden',
      clip: 'rect(0 0 0 0)',
      height: 1,
      width: 1,
      margin: -1,
    },
  },
});


class Option extends Component {
  handleClick = event => {
    this.props.onSelect(this.props.option, event);
  };

  render() {
    const { children, isFocused, isSelected, onFocus } = this.props;

    return (
      <MenuItem
        onFocus={onFocus}
        selected={isFocused}
        onClick={this.handleClick}
        component="div"
        style={{
          fontWeight: isSelected ? 500 : 400,
        }}
      >
        {children}
      </MenuItem>
    );
  }
}

function SelectWrapped(props) {
  const { classes, ...other } = props;

  return (
    <Select.Async
      optionComponent={Option}
      noResultsText={<Typography>{'No results found'}</Typography>}
      arrowRenderer={arrowProps => {
        return arrowProps.isOpen ? <MenuUp /> : <MenuDown />;
      }}
      clearRenderer={() => <Close />}
      valueComponent={valueProps => {
        const { value, children, onRemove } = valueProps;

        const onDelete = event => {
          event.preventDefault();
          event.stopPropagation();
          onRemove(value);
        };

        if (onRemove) {
          return (
            <Chip
              tabIndex={-1}
              label={children}
              className={classes.chip}
              deleteIcon={<Cancel onTouchEnd={onDelete} />}
              onDelete={onDelete}
            />
          );
        }

        return <div className="Select-value">{children}</div>;
      }}
      {...other}
    />
  );
}


class AutocompleteSelect extends Component {
  constructor() {
    super();

    this.state = {
      options: [],
    };

    this.setInitialOptions = this.setInitialOptions.bind(this);
    this.setSelectedOption = this.setSelectedOption.bind(this);
    this.onAutocomplete = this.onAutocomplete.bind(this);
    this.onChange = this.onChange.bind(this);
  }

  componentDidMount() {
    this.setInitialOptions(this.setSelectedOption);
  }

  componentDidUpdate(prevProps) {
    if (prevProps.value === this.props.value && prevProps.triggerReload === this.props.triggerReload) {
      return;
    }

    this.setInitialOptions(this.setSelectedOption);
  }

  setInitialOptions(callbackFunc) {
    this.props.getOptions("", options => {
      this.setState({
        options: options,
      }, callbackFunc);
    });
  }

  setSelectedOption() {
    if (this.props.getOption !== undefined) {
      if (this.props.value !== undefined && this.props.value !== "" && this.props.value !== null) {
        this.props.getOption(this.props.value, resp => {
          this.setState({
            selectedOption: resp,
          });
        });
      } else {
        this.setState({
          selectedOption: "",
        });
      }
    } else {
      if (this.props.value !== undefined && this.props.value !== "" && this.props.value !== null) {
        for (const opt of this.state.options) {
          if (this.props.value === opt.value) {
            this.setState({
              selectedOption: opt,
            });
          }
        }
      } else {
        this.setState({
          selectedOption: "",
        });
      }
    }
  }

  onAutocomplete(input, callbackFunc) {
    this.props.getOptions(input, options => {
      this.setState({
        options: options,
      });

      callbackFunc(null, {
        options: options,
        complete: true,
      });
    });
  }

  onChange(v) {
    let value = null;
    if (v !== null) {
      value = v.value;
    }

    this.setState({
      selectedOption: v,
    });

    this.props.onChange({
      target: {
        id: this.props.id,
        value: value,
      },
    });
  }

  render() {
    const inputProps = this.props.inputProps || {};
    return(
      <FormControl margin={this.props.margin || ""}  fullWidth={true} className={this.props.className}>
        <Input
          fullWidth
          inputComponent={SelectWrapped}
          placeholder={this.props.label}
          id={this.props.id}
          onChange={this.onChange}
          inputProps={{...{
            instanceId: this.props.id,
            clearable: false,
            options: this.state.options,
            loadOptions: this.onAutocomplete,
            value: this.state.selectedOption || "",
          }, ...inputProps}}
        />
      </FormControl>
    );
  }
}

export default withStyles(styles)(AutocompleteSelect);
