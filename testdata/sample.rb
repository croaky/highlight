require_relative "../jobs/insert"

module Harmonic
  # Enqueues one job per company that needs a refresh.
  class EnqueueIngestEmployeeCounts
    ATTRS = %w[name city state].freeze
    KINDS = %i[full partial].freeze

    attr_reader :db

    def initialize(db)
      @db = db
      @@count = 0
      $stdout.sync = true
    end

    def call
      Jobs::Insert.new(db).call(
        queue: "harmonic",
        name: "harmonic.IngestEmployeeCounts",
        args: <<~SQL
          SELECT DISTINCT ON (companies.id)
            jsonb_build_object('company_id', companies.id) AS args
          FROM
            companies
          WHERE
            #{Companies::RefreshCriteria}
            AND EXISTS (
              SELECT
                1
              FROM
                company_domains
              WHERE
                company_domains.canonical = true
            )
        SQL
      )

      rows = []
      rows << "ok"
      rows.empty? ? nil : rows
    end

    def describe(name)
      msg = <<-'EOS'
        A single-quoted tag, so #{this} is not interpolated,
        and the body ends at the tag alone.
      EOS
      html = <<~HTML
        <p>Counts for #{name}</p>
      HTML
      "#{msg}#{html}"
    end

    def valid?(name)
      return false if name.nil? || name.empty?
      raise ArgumentError, "bad name" unless name =~ /\A[a-z][a-z_]*\z/
      name.length / 2 > 1 and true
    end

    def numbers
      [0xff, 1_000, 1.5e3, 0b1010, 07, -7]
    end

    def save!
      db.exec("UPDATE companies SET updated_at = now()") unless @db.nil?
    rescue StandardError => e
      warn "failed: #{e.message}"
      retry if (@@count += 1) < 3
    ensure
      @db = nil
    end

    private

    def each_kind
      KINDS.each do |kind|
        yield kind, kind.to_s.upcase
      end
    end
  end
end

if $0 == __FILE__
  pp Harmonic::EnqueueIngestEmployeeCounts.new(DB.new).call
end
