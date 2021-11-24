import { Field } from "react-final-form";
import React from "react";
import Error from "./Error";
import { classNames } from "../../styles/utils";

type COL_WIDTHS = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;

interface Props {
    name: string;
    label?: string;
    placeholder?: string;
    columns?: COL_WIDTHS;
    className?: string;
    autoComplete?: string;
}

const PasswordField: React.FC<Props> = ({ name, label, placeholder, columns, className, autoComplete }) => (
    <div
        className={classNames(
            columns ? `col-span-${columns}` : "col-span-12"
        )}
    >
        {label && (
            <label htmlFor={name} className="block text-xs font-bold text-gray-700 dark:text-white uppercase tracking-wide">
                {label}
            </label>
        )}
        <Field
            name={name}
            render={({ input }) => (
                <input
                    {...input}
                    id={name}
                    type="password"
                    autoComplete={autoComplete}
                    className="mt-2 block w-full border border-gray-300 dark:border-gray-700 dark:bg-gray-800 dark:text-white rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    placeholder={placeholder}
                />
            )}
        />
        <div>
            <Error name={name} classNames="text-red mt-2" />
        </div>
    </div>
)

export default PasswordField;