class FixedSizeSlidingCounter < Array
  def initialize(max_size)
    @max_size = max_size
    @counts = Hash.new(0)
  end

  def increment(key)
    if size >= @max_size
      @counts[shift] -= 1
    end

    push(key)
    @counts[key] += 1
  end

  def ratios
    ratios = {}
    @counts.each do |priority, count|
      ratios[priority] = count.to_f/self.size
    end
    ratios
  end
end

class Throttler
  def initialize(steps)
    @steps = steps
    @mod = 0
  end

  def allow(drop_ratio)
    return false if drop_ratio >= 1
    return true if drop_ratio <= 0
    #
    # the @mod variable slides over the following structure rejecting/accepting requests
    #
    # 0                                             steps
    # -------------------------------------------------
    # |R|R|R|R|R|R|R|R|R|R|R|R|R|R|A|A|A|A|A|A|A|A|A|A|
    # -------------------------------------------------
    #                             ^thresh
    @mod = (@mod + 1) % @steps
    threshold = @steps - drop_ratio * @steps
    threshold > @mod
  end
end

class Controller
  DEFAULT_PRIORITY = :default

  # in order of increasing priority
  PRIORITIES = [
    :offender,
    DEFAULT_PRIORITY
  ].freeze

  Scope = Struct.new(:pod_id)

  def initialize(soft, hard)
    @soft_limit = soft
    @hard_limit = hard
    @drop_ratios = {}
    @scope_priorities = {} # { scope => priority }
    @throttlers = {} # { priority => throttler }
    @frequency_counter = FixedSizeSlidingCounter.new(1000)

    PRIORITIES.each do |priority|
      @throttlers[priority] = Throttler.new(100)
    end
  end

  def allow(req_scope, load_val)
    update_priority_frequency(req_scope)
    update_drop_ratios(load_val)
    drop_request?(req_scope)
  end

  def add_scope(scope, priority: :default)
    @scope_priorities[scope] = priority
  end

  private

  def drop_request?(req_scope)
    req_priority = @scope_priorities[req_scope] || :default
    @throttlers[req_priority]&.allow(@drop_ratios[req_priority] || 0)
  end

  def update_priority_frequency(req_scope)
    req_priority = @scope_priorities[req_scope] || :default
    @frequency_counter.increment(req_priority)
  end

  def update_drop_ratios(load_val)
    # drop rate is the following ratio:
    #
    # num_rejected / num_requests
    #
    # for example, a ratio of 1/4 means for every 4 requests drop 1 one of them
    # in other words, drop 25% of requests

    # calculate the target global drop rate based on a linear scale:
    #
    #     soft                   hard
    #  <---|-----------|-----------|--->
    #      0%         50%         100%
    #
    length = (@hard_limit - @soft_limit)
    target_drop_rate = 1 - ((@hard_limit - load_val) / length)

    # given the new load value and the observed priority frequencies, work
    # backwards from the desired global drop rate to a set of ratios for active
    # throttlers
    target_sum = target_drop_rate
    PRIORITIES.each do |priority|
      # we've already reached the target rate, pass all other priorities
      if target_sum == 0
        @drop_ratios[priority] = 0
        next
      end

      # expected percentage of total load that can be shed in this priority bucket
      priority_frequency = @frequency_counter.ratios[priority] || 0

      if target_sum <= priority_frequency
        # calculate how much of the last priority bucket must be shed
        @drop_ratios[priority] = target_sum.to_f / priority_frequency
        target_sum -= target_sum
      else
        # must reject 100% of requests from the current priority in order to
        # protect the next priority. we essentially downgrade priority at this
        # point
        @drop_ratios[priority] = 1
        target_sum -= priority_frequency
      end
    end
  end
end

load_range = (70..110)
soft = 80.0
hard = 100.0
total_requests_per_load = 5000

load_range.each do |current_load|
  scopes = [:a, :a, :a, :a, :a, :a, :b, :b, :b, :b, :c, :d, :e]

  allowed = {}
  scopes.each do |scope|
    allowed[scope] = {true => 0, false => 0}
  end

  t = Controller.new(soft, hard)
  t.add_scope(:a, priority: :offender)

  passed_requests = 0

  (0...total_requests_per_load).each do |i|
    scope = scopes.sample
    pass = t.allow(scope, current_load)
    allowed[scope][pass] += 1
    passed_requests += 1 if pass
  end

  length = (hard - soft)
  target_ratio = (hard - current_load) / length

  puts "passed_requests=#{passed_requests} pass_ratio=#{passed_requests.to_f/total_requests_per_load} target_ratio=#{target_ratio} allowed=#{allowed}"
  scopes.uniq.each do |scope|
    sum = allowed[scope][true] + allowed[scope][false]
    ratio = allowed[scope][true].to_f / sum.to_f
    puts "scope=#{scope} load=#{current_load} pass_ratio=#{ratio.round(2)}\t\t"
  end
  puts
end
