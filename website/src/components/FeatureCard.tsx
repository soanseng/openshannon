interface FeatureCardProps {
  icon: string
  title: string
  description: string
}

export default function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div className="bg-white border border-card-border rounded-xl p-6 hover:border-accent/40 transition-colors">
      <div className="text-3xl mb-3">{icon}</div>
      <h3 className="font-semibold text-lg text-navy mb-2">{title}</h3>
      <p className="text-navy-light text-sm leading-relaxed">{description}</p>
    </div>
  )
}
