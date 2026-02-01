"use client";

import dynamic from "next/dynamic";
import type { ReactCodeMirrorProps } from "@uiw/react-codemirror";

const CodeMirror = dynamic(
  () => import("@uiw/react-codemirror").then((m) => m.default),
  { ssr: false },
);

export function CodeEditor(props: {
  value: string;
  height: string;
  extensions: NonNullable<ReactCodeMirrorProps["extensions"]>;
  editable?: boolean;
  onChange: (value: string) => void;
  theme?: ReactCodeMirrorProps["theme"];
}) {
  return (
    <CodeMirror
      value={props.value}
      height={props.height}
      extensions={props.extensions}
      theme={props.theme}
      onChange={props.onChange}
      editable={props.editable}
    />
  );
}

