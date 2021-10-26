import React from "react";
import { classNames } from "./Utils";

export function ButtonGroup({ children, className, ...rest }) {
  return (
    <div
      role="group"
      className={classNames(
        "flex rounded-sm text-sm",
        className.includes("justify-") ? "" : "justify-center",
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  );
}

export function GroupButton({ children, className, dir, ...rest }) {
  return (
    <button
      className={classNames(
        "border border-blue-500 px-4 py-2 mx-0 outline-none focus:shadow-outline",
        rest.left
          ? "border-r-0 rounded-l-sm"
          : rest.right
          ? "border-l-0 rounded-r-sm"
          : "border-r-0 border-l-0",
        rest.active ? "bg-blue-500 text-white" : "bg-white text-blue-500",
        className
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
