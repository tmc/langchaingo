/* eslint-disable react/jsx-props-no-spreading */
import React from "react";
import CodeBlock from "@theme-original/CodeBlock";

export default function CodeBlockWrapper({ children, ...props }) {
  if (typeof children === "string") {
    return <CodeBlock {...props}>{children}</CodeBlock>;
  }
  console.log(children)
  return (
    <>
      <CodeBlock {...props}>{children.content}</CodeBlock>
    </>
  );
}
