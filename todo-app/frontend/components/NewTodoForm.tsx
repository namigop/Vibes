"use client";

import { useForm } from "react-hook-form";
import { api } from "@/lib/api";

interface FormShape {
  title: string;
  description?: string;
}

interface Props {
  onCreated: () => void;
}

export function NewTodoForm({ onCreated }: Props) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormShape>({ defaultValues: { title: "", description: "" } });

  const onSubmit = handleSubmit(async (values) => {
    await api.createTodo({
      title: values.title.trim(),
      description: values.description?.trim() ?? "",
    });
    reset({ title: "", description: "" });
    onCreated();
  });

  return (
    <form
      onSubmit={onSubmit}
      className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-800"
    >
      <div className="flex flex-col gap-2">
        <label className="text-sm font-medium" htmlFor="title">
          New todo
        </label>
        <input
          id="title"
          type="text"
          autoComplete="off"
          placeholder="What needs doing?"
          className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-900"
          {...register("title", {
            required: "Title is required",
            maxLength: { value: 200, message: "Title is too long (max 200)" },
          })}
        />
        {errors.title ? (
          <p className="text-xs text-red-600">{errors.title.message}</p>
        ) : null}

        <textarea
          id="description"
          rows={2}
          placeholder="Optional description"
          className="w-full resize-none rounded-md border border-slate-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-900"
          {...register("description", {
            maxLength: { value: 2000, message: "Description is too long (max 2000)" },
          })}
        />
        {errors.description ? (
          <p className="text-xs text-red-600">{errors.description.message}</p>
        ) : null}

        <button
          type="submit"
          disabled={isSubmitting}
          className="self-end rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {isSubmitting ? "Adding…" : "Add todo"}
        </button>
      </div>
    </form>
  );
}
