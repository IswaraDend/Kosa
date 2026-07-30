interface BadgeProps {
  label: string;
  variant: string;
}

const Badge = ({ label, variant }: BadgeProps) => {
  return <span className={`badge ${variant}`}>{label}</span>;
};

export default Badge;
