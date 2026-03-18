import { createContext, useContext, type ReactNode } from "react";
import {
  Controller,
  FormProvider,
  useFormContext,
  type ControllerProps,
  type FieldPath,
  type FieldValues,
  type UseFormReturn,
} from "react-hook-form";

type FormFieldContextValue = {
  name: string;
};

const FormFieldContext = createContext<FormFieldContextValue | null>(null);

export function Form<TFieldValues extends FieldValues>({ children, ...props }: UseFormReturn<TFieldValues> & { children: ReactNode }) {
  return <FormProvider {...props}>{children}</FormProvider>;
}

export function FormField<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>(
  props: ControllerProps<TFieldValues, TName>,
) {
  return (
    <FormFieldContext.Provider value={{ name: props.name }}>
      <Controller {...props} />
    </FormFieldContext.Provider>
  );
}

export function FormItem({ children }: { children: ReactNode }) {
  return <div className="space-y-2">{children}</div>;
}

export function FormLabel({ children, htmlFor }: { children: ReactNode; htmlFor?: string }) {
  return (
    <label className="block text-sm font-medium text-[var(--ink)]" htmlFor={htmlFor}>
      {children}
    </label>
  );
}

export function FormControl({ children }: { children: ReactNode }) {
  return <div>{children}</div>;
}

export function FormDescription({ children }: { children: ReactNode }) {
  return <p className="text-xs leading-6 text-[var(--muted)]">{children}</p>;
}

export function FormMessage() {
  const fieldContext = useContext(FormFieldContext);
  const formContext = useFormContext();

  if (!fieldContext) {
    return null;
  }

  const error = formContext.formState.errors[fieldContext.name];

  if (!error || typeof error.message !== "string") {
    return null;
  }

  return <p className="text-sm font-medium text-red-700">{error.message}</p>;
}
