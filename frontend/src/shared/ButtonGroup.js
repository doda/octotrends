import React from "react";
import { classNames } from "./Utils";

export function ButtonGroup({ children, className, ...rest }) {
  return (
    <div
      role="group"
      className={classNames(
        "flex rounded-sm text-sm",
        className.includes("justify-") ? "" : "justify-center",
        className
      )}
      {...rest}
    >
      {children}
    </div>
  );
}

export function GroupButton({
  children,
  className,
  dir,
  left,
  right,
  active,
  ...rest
}) {
  return (
    <button
      className={classNames(
        "border border-blue-500 lg:px-4 px-2 py-2 mx-0 outline-none focus:shadow-outline",
        left
          ? "border-r-0 rounded-l-sm"
          : right
          ? "border-l-0 rounded-r-sm"
          : "border-r-0 border-l-0",
        active ? "bg-blue-500 text-white" : "bg-white text-blue-500",
        className
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
