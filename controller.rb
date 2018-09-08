class Test
  def initialize(soft, hard, steps, silent: false)
    @mod = 0
    @soft = soft
    @hard = hard
    @steps = steps
    @silent = silent
  end

  def allow(load_val)
    divisor = (@hard - @soft) / @steps

    @mod = (@mod + 1) % @steps

    threshold = (@hard - load_val) / divisor

    pass = threshold > @mod

    puts "load=#{load_val}\t\tthreshold=#{threshold}\t@mod=#{@mod}\tpass=#{pass}" unless @silent

    pass
  end
end

class Throttler
  def initialize(steps)
    @steps = steps
    @mod = 0
  end

  def allow(drop_ratio)
    pass_ratio = 1 - drop_ratio

    return false if pass_ratio <= 0
    return true if pass_ratio >= 1

    @mod = (@mod + 1) % @steps

    threshold = @steps - pass_ratio * @steps

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

    PRIORITIES.each do |priority|
      @throttlers[priority] = Throttler.new(100)
    end
  end

  def allow(req_scope, load_val)
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

  def update_drop_ratios(load_val)
    # drop rate is the following ratio:
    #
    # num_rejected / num_requests
    #
    #
    # for example, a ratio of 1/4 means for every 4 requests drop 1 one of them
    # in other words, drop 25% of requests
    #
    # calculate the target _global_ drop rate:
    length = (@hard_limit - @soft_limit)
    target_drop_rate = (@hard_limit - load_val) / length
    target_drop_rate = 1 - target_drop_rate

    num_throttlers = @throttlers.size

    # given the new load value, work backwards from the desired global drop
    # rate to a set of ratios for active throttlers
    #
    # i.e. calculate the sum of the distribution using its expected value and
    # then re-distribute the sum, biasing towards the lowest priority
    target_sum = target_drop_rate * num_throttlers
    priority = 0

    while priority < num_throttlers && target_sum > 0
      if target_sum <= 1
        # last iteration of the loop. sum of all the drop ratios equals the
        # initial value of target_sum
        @drop_ratios[PRIORITIES[priority]] = target_sum
        target_sum -= 1
      else
        # must reject 100% of requests from the current priority in order to
        # protect the next priority. we essentially downgrade priority at this
        # point
        @drop_ratios[PRIORITIES[priority]] = 1
        target_sum -= 1
        priority += 1
      end
    end
  end
end

allowed = {true => 0, false => 0}
t = Test.new(80.0, 100.0, 5.0)
(0..1000).each do |i|
  allowed[t.allow(i.to_f)] += 1
end
puts "allowed=#{allowed}"



(60..110).each do |load_|
  allowed = {true => 0, false => 0}

  t = Test.new(80.0, 100.0, 100.0, silent: true)
  (0..200).each do |i|
    allowed[t.allow(load_)] += 1
  end

  sum = allowed[true] + allowed[false]
  ratio = allowed[true].to_f / sum.to_f
  puts "load=#{load_} allowed=#{allowed} ratio=#{ratio}"
end


(80..93).each do |load_|
  scopes = [:b, :b, :b, :b, :c, :d, :e]

  allowed = {}
  scopes.each do |scope|
    allowed[scope] = {true => 0, false => 0}
  end

  soft = 80.0
  hard = 100.0

  t = Controller.new(soft, hard)
  t.add_scope(:a, priority: :offender)

  passed_requests = 0
  total_requests = 5000

  (0...total_requests).each do |i|
    scope = scopes.sample
    pass = t.allow(scope, load_)
    allowed[scope][pass] += 1
    passed_requests += 1 if pass
  end

  length = (hard - soft)
  target_ratio = (hard - load_) / length

  puts "passed_requests=#{passed_requests} pass_ratio=#{passed_requests.to_f/total_requests} target_ratio=#{target_ratio} allowed=#{allowed}"

  scopes.uniq.each do |scope|
    sum = allowed[scope][true] + allowed[scope][false]
    ratio = allowed[scope][true].to_f / sum.to_f
    puts "scope=#{scope} load=#{load_} pass_ratio=#{ratio.round(2)}\t\t"
  end
  puts
end
