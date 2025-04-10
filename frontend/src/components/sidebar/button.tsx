type ButtonProps = {
    children: React.ReactNode
    onClick: () => void
}

export const Button = ({ children, onClick }: ButtonProps) => {
    return (
        <button onClick={onClick} className="p-3 text-white hover:bg-[#383838] rounded-xl transition-colors cursor-pointer">
            {children}
        </button>
    )
}